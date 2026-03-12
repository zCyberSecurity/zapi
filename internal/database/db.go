package database

import (
	"log"

	"github.com/zCyberSecurity/zapi/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(dbPath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Provider{},
		&model.ProviderModel{},
		&model.APIKey{},
		&model.UsageLog{},
	); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	return db
}
