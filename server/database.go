package server

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/websentry/websentry/config"
)

func connectToDB(dbConfig config.Database) (*gorm.DB, error) {
	switch dbConfig.Type {
	case "postgres":
		return gorm.Open(postgres.Open(dbConfig.DataSourceName), &gorm.Config{})
	default:
		return nil, fmt.Errorf("Unsupported database type: %v", dbConfig.Type)
	}
}
