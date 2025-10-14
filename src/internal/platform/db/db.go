package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pocketeer/internal/config"
	"pocketeer/internal/platform/db/models"
)

type DB = gorm.DB;

func Init(cfg *config.Config) *DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DbHost,
		cfg.DbUser,
		cfg.DbPassword,
		cfg.DbName,
		cfg.DbPort,
		cfg.DbSSLMode,
		"UTC",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Set underlying connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// NOTE(pencelheimer/gemini): consider dedicated migration tools (Goose, Migrate)
	err = db.AutoMigrate(
		&models.User{},
	)
	if err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("Database connection successfully initialized and migrated.")

	return db
}
