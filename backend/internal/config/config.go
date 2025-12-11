package config

import (
	"github.com/pgmod/envconfig"
)

type Config struct {
	Port        string
	DbHost      string
	DbPort      string
	DbName      string
	DbUser      string
	DbPassword  string
	NeedMigrate bool
}

func ReadConfig() (*Config, error) {
	if err := envconfig.Load(); err != nil {
		return nil, err
	}

	port := envconfig.Get("PORT", "8080")
	dbHost := envconfig.Get("DB_HOST", "localhost")
	dbPort := envconfig.Get("DB_PORT", "5432")
	dbName := envconfig.Get("POSTGRES_DB", "your_database")
	dbUser := envconfig.Get("POSTGRES_USER", "postgres")
	dbPassword := envconfig.Get("POSTGRES_PASSWORD", "postgres")
	needMigrate := envconfig.GetBool("NEED_MIGRATE", true)
	return &Config{
		Port:        port,
		DbHost:      dbHost,
		DbPort:      dbPort,
		DbName:      dbName,
		DbUser:      dbUser,
		DbPassword:  dbPassword,
		NeedMigrate: needMigrate,
	}, nil
}
