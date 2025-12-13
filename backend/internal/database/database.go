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

	log.Println("Database schema initialized successfully")
	return nil
}
