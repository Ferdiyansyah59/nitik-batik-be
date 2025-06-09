package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func SetupDatabaseConnection() *gorm.DB {
	// errEnv := godotenv.Load()
	// if errEnv != nil {
	// 	panic("Failed to load env file")
	// }
	_ = godotenv.Load()

	dbUser := os.Getenv("DB_USERNAME")
	dbPw := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_DATABASE")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3307)/%s?charset=utf8&parseTime=True&loc=Local", dbUser, dbPw, dbHost, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Failed to Create a connection to database")
	}

	return db
}

// Close Database
func CloseDatabaseConnection(db *gorm.DB) {
	dbSQL, err := db.DB()

	if err != nil {
		panic("Failed to close connection from database")
	}

	dbSQL.Close()
}
