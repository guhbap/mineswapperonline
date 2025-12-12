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
	// AutoMigrate создаст таблицы, если их нет, и обновит существующие
	err := db.AutoMigrate(
		&models.User{},
		&models.UserStats{},
		&models.UserBestResult{},
		&models.UserGameHistory{},
		&models.GameParticipant{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Создаем индексы вручную, так как GORM не всегда корректно их создает
	err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_user_game_history_user_id_rating_gain 
		ON user_game_history(user_id, rating_gain DESC);
	`).Error
	if err != nil {
		log.Printf("Warning: failed to create index idx_user_game_history_user_id_rating_gain: %v", err)
	}

	err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_game_participants_game_history_id 
		ON game_participants(game_history_id);
	`).Error
	if err != nil {
		log.Printf("Warning: failed to create index idx_game_participants_game_history_id: %v", err)
	}

	log.Println("Database schema initialized successfully")
	return nil
}
