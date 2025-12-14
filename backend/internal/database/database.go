package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"minesweeperonline/internal/models"
)

type DB struct {
	*gorm.DB
}

func NewDB(host, port, user, password, dbname string) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")

	return &DB{db}, nil
}

func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *DB) InitSchema() error {
	migrator := db.Migrator()

	// Проверяем существование таблиц и создаем их, если нужно
	if !migrator.HasTable(&models.User{}) {
		if err := migrator.CreateTable(&models.User{}); err != nil {
			return fmt.Errorf("failed to create users table: %w", err)
		}
	} else {
		// Если таблица существует, обновляем схему через AutoMigrate для добавления новых полей
		if err := migrator.AutoMigrate(&models.User{}); err != nil {
			log.Printf("Warning during migration of User: %v", err)
		}
		// Создаем уникальные ограничения вручную, если их нет
		db.Exec(`
			DO $$ 
			BEGIN
				IF NOT EXISTS (
					SELECT 1 FROM pg_constraint WHERE conname = 'users_username_key'
				) THEN
					ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);
				END IF;
			END $$;
		`)

		db.Exec(`
			DO $$ 
			BEGIN
				IF NOT EXISTS (
					SELECT 1 FROM pg_constraint WHERE conname = 'users_email_key'
				) THEN
					ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
				END IF;
			END $$;
		`)
	}

	// Создаем остальные таблицы через AutoMigrate
	// Используем отдельные вызовы для лучшего контроля ошибок
	tables := []interface{}{
		&models.UserStats{},
		&models.UserGameHistory{},
		&models.GameParticipant{},
		&models.Room{},
	}

	for _, table := range tables {
		if err := migrator.AutoMigrate(table); err != nil {
			log.Printf("Warning during migration of %T: %v", table, err)
			// Продолжаем выполнение, так как таблицы могут быть уже созданы
		}
	}

	// Явно добавляем новые поля в user_game_history, если их нет
	if migrator.HasTable(&models.UserGameHistory{}) {
		// Проверяем и добавляем room_id
		if !migrator.HasColumn(&models.UserGameHistory{}, "room_id") {
			if err := db.Exec(`ALTER TABLE user_game_history ADD COLUMN room_id VARCHAR(255)`).Error; err != nil {
				log.Printf("Warning: failed to add room_id column: %v", err)
			} else {
				log.Println("Added room_id column to user_game_history")
			}
		}
		// Проверяем и добавляем seed (UUID)
		if !migrator.HasColumn(&models.UserGameHistory{}, "seed") {
			if err := db.Exec(`ALTER TABLE user_game_history ADD COLUMN seed VARCHAR(36) NOT NULL DEFAULT ''`).Error; err != nil {
				log.Printf("Warning: failed to add seed column: %v", err)
			} else {
				log.Println("Added seed column to user_game_history")
			}
		} else {
			// Если колонка существует, проверяем тип и изменяем если нужно
			var columnInfo struct {
				DataType      string
				CharMaxLength *int
			}
			if err := db.Raw(`
				SELECT data_type, character_maximum_length 
				FROM information_schema.columns 
				WHERE table_name = 'user_game_history' AND column_name = 'seed'
			`).Scan(&columnInfo).Error; err == nil {
				needsConversion := false
				if columnInfo.DataType == "bigint" || columnInfo.DataType == "integer" || columnInfo.DataType == "numeric" {
					needsConversion = true
					log.Printf("Converting seed column from %s to VARCHAR(36)", columnInfo.DataType)
				} else if columnInfo.DataType == "character varying" || columnInfo.DataType == "varchar" {
					// Проверяем размер - если меньше 36, нужно изменить
					if columnInfo.CharMaxLength == nil || *columnInfo.CharMaxLength < 36 {
						log.Printf("Seed column size is %v, need to change to VARCHAR(36)", columnInfo.CharMaxLength)
						if err := db.Exec(`ALTER TABLE user_game_history ALTER COLUMN seed TYPE VARCHAR(36)`).Error; err != nil {
							log.Printf("Warning: failed to change seed column size: %v", err)
						} else {
							log.Println("Changed seed column size to VARCHAR(36)")
						}
					}
				}

				if needsConversion {
					// В PostgreSQL нельзя напрямую конвертировать BIGINT в VARCHAR
					// Используем временную колонку для миграции
					// Сначала проверяем, нет ли уже временной колонки
					var tempColumnExists bool
					db.Raw(`SELECT EXISTS (
						SELECT 1 FROM information_schema.columns 
						WHERE table_name = 'user_game_history' AND column_name = 'seed_new'
					)`).Scan(&tempColumnExists)

					if !tempColumnExists {
						// Создаем временную колонку
						if err := db.Exec(`ALTER TABLE user_game_history ADD COLUMN seed_new VARCHAR(36) DEFAULT ''`).Error; err != nil {
							log.Printf("Warning: failed to add temporary seed column: %v", err)
						} else {
							log.Println("Added temporary seed_new column")
							// Генерируем UUID для всех существующих записей
							if err := db.Exec(`
								UPDATE user_game_history 
								SET seed_new = LOWER(
									SUBSTRING(MD5(RANDOM()::TEXT || CLOCK_TIMESTAMP()::TEXT || id::TEXT) FROM 1 FOR 8) || '-' ||
									SUBSTRING(MD5(RANDOM()::TEXT || CLOCK_TIMESTAMP()::TEXT || id::TEXT) FROM 9 FOR 4) || '-' ||
									'4' || SUBSTRING(MD5(RANDOM()::TEXT || CLOCK_TIMESTAMP()::TEXT || id::TEXT) FROM 14 FOR 3) || '-' ||
									SUBSTRING('89ab', FLOOR(RANDOM() * 4 + 1)::INT, 1) || SUBSTRING(MD5(RANDOM()::TEXT || CLOCK_TIMESTAMP()::TEXT || id::TEXT) FROM 18 FOR 3) || '-' ||
									SUBSTRING(MD5(RANDOM()::TEXT || CLOCK_TIMESTAMP()::TEXT || id::TEXT) FROM 22 FOR 12)
								)
								WHERE seed_new = '' OR seed_new IS NULL
							`).Error; err != nil {
								log.Printf("Warning: failed to generate UUIDs for existing records: %v", err)
							} else {
								log.Println("Generated UUIDs for existing records")
							}
							// Удаляем старую колонку
							if err := db.Exec(`ALTER TABLE user_game_history DROP COLUMN seed`).Error; err != nil {
								log.Printf("Warning: failed to drop old seed column: %v", err)
							} else {
								log.Println("Dropped old seed column")
							}
							// Переименовываем новую колонку
							if err := db.Exec(`ALTER TABLE user_game_history RENAME COLUMN seed_new TO seed`).Error; err != nil {
								log.Printf("Warning: failed to rename seed_new column: %v", err)
							} else {
								log.Println("Renamed seed_new to seed")
							}
							// Устанавливаем NOT NULL после заполнения данных
							if err := db.Exec(`ALTER TABLE user_game_history ALTER COLUMN seed SET NOT NULL`).Error; err != nil {
								log.Printf("Warning: failed to set seed NOT NULL: %v", err)
							} else {
								log.Println("Set seed column to NOT NULL")
							}
							log.Println("Converted seed column from BIGINT to VARCHAR(36)")
						}
					} else {
						log.Println("Temporary seed_new column already exists, skipping conversion")
					}
				}
			}
		}
		// Проверяем и добавляем creator_id
		if !migrator.HasColumn(&models.UserGameHistory{}, "creator_id") {
			if err := db.Exec(`ALTER TABLE user_game_history ADD COLUMN creator_id INTEGER NOT NULL DEFAULT 0`).Error; err != nil {
				log.Printf("Warning: failed to add creator_id column: %v", err)
			} else {
				log.Println("Added creator_id column to user_game_history")
			}
		}
		// Проверяем и добавляем won
		if !migrator.HasColumn(&models.UserGameHistory{}, "won") {
			if err := db.Exec(`ALTER TABLE user_game_history ADD COLUMN won BOOLEAN NOT NULL DEFAULT false`).Error; err != nil {
				log.Printf("Warning: failed to add won column: %v", err)
			} else {
				log.Println("Added won column to user_game_history")
			}
		}
		// Проверяем и добавляем has_custom_seed
		if !migrator.HasColumn(&models.UserGameHistory{}, "has_custom_seed") {
			if err := db.Exec(`ALTER TABLE user_game_history ADD COLUMN has_custom_seed BOOLEAN NOT NULL DEFAULT false`).Error; err != nil {
				log.Printf("Warning: failed to add has_custom_seed column: %v", err)
			} else {
				log.Println("Added has_custom_seed column to user_game_history")
			}
		}
	}

	// Создаем уникальные индексы вручную, если их нет
	db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
	`)
	db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`)

	err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_game_participants_game_history_id 
		ON game_participants(game_history_id);
	`).Error
	if err != nil {
		log.Printf("Warning: failed to create index idx_game_participants_game_history_id: %v", err)
	}

	// Удаляем поле rating из таблицы users, если оно существует (рейтинг теперь рассчитывается динамически)
	if migrator.HasTable(&models.User{}) {
		if migrator.HasColumn(&models.User{}, "rating") {
			if err := db.Exec(`ALTER TABLE users DROP COLUMN IF EXISTS rating`).Error; err != nil {
				log.Printf("Warning: failed to drop rating column: %v", err)
			} else {
				log.Println("Dropped rating column from users table (rating is now calculated dynamically)")
			}
		}
	}

	log.Println("Database schema initialized successfully")
	return nil
}
