package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(host, port, user, password, dbname string) (*DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")

	return &DB{db}, nil
}

func (db *DB) InitSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			color VARCHAR(7) DEFAULT NULL,
			rating DOUBLE PRECISION DEFAULT 1500.0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS color VARCHAR(7) DEFAULT NULL;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS rating DOUBLE PRECISION DEFAULT 1500.0;`,
		`CREATE TABLE IF NOT EXISTS user_stats (
			user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			games_played INTEGER DEFAULT 0,
			games_won INTEGER DEFAULT 0,
			games_lost INTEGER DEFAULT 0,
			last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS user_sessions (
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			session_token VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			PRIMARY KEY (user_id, session_token)
		);`,
		`CREATE TABLE IF NOT EXISTS user_best_results (
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			width INTEGER NOT NULL,
			height INTEGER NOT NULL,
			mines INTEGER NOT NULL,
			best_time DOUBLE PRECISION NOT NULL,
			complexity DOUBLE PRECISION NOT NULL,
			best_p DOUBLE PRECISION DEFAULT 0.0,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, width, height, mines)
		);`,
		`ALTER TABLE user_best_results ADD COLUMN IF NOT EXISTS best_p DOUBLE PRECISION DEFAULT 0.0;`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	log.Println("Database schema initialized successfully")
	return nil
}

