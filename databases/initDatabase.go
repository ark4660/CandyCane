package database

import (
	"gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)


func InitDatabase() *gorm.DB {
	dsn := "host=localhost user=postgres password=pHAN4002@p dbname=mydb port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    	Logger: logger.Default.LogMode(logger.Info), // shows all SQL queries
    })
    if err != nil {
        panic("failed to connect to database: " + err.Error())
    }

    // Auto-migrate tables
    err = db.AutoMigrate(&VideoInformation{})
    if err != nil {
        panic("failed to migrate database: " + err.Error())
    }
    return db
}

func InitDatabaseSessionHistory() *gorm.DB {
	dsn := "host=localhost user=postgres password=pHAN4002@p dbname=mydb port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    	Logger: logger.Default.LogMode(logger.Info), // shows all SQL queries
    })
    if err != nil {
        panic("failed to connect to database: " + err.Error())
    }

    // Auto-migrate tables
    err = db.AutoMigrate(&VideoSession{})
    if err != nil {
        panic("failed to migrate database: " + err.Error())
    }
    return db
}
