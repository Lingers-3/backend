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

type DB = gorm.DB

func Init(cfg *config.Config) (db *DB, err error) {
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

	sleep := time.Duration(500 * time.Millisecond)
	attempts := 5
	for i := range attempts {
		log.Println("Establishing connection to the DB. Attempt ", i+1)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Printf("Error connecting to the DB: %s. Sleeping for: %s", err, sleep)
			time.Sleep(sleep)
			sleep *= 2
			continue
		}
		break
	}
	if err != nil {
		return nil, err
	}

	// Set underlying connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("Failed to get underlying sql.DB: %v", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// NOTE(pencelheimer/gemini): consider dedicated migration tools (Goose, Migrate)
	err = db.AutoMigrate(
		&models.User{},
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to run migrations: %v", err)
	}

	log.Println("Database connection successfully initialized and migrated.")

	return db, nil
}
