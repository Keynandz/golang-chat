package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

func InitDB() {
	loadErr := godotenv.Load()
	if loadErr != nil {
		log.Fatal("error loading file .env")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHost, dbUser, dbPass, dbName, dbPort)
	var gormErr error
	db, gormErr = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if gormErr != nil {
		panic(gormErr.Error())
	}

	sqlDB, sqlDBErr := db.DB()
	if sqlDBErr != nil {
		panic(sqlDBErr.Error())
	}

	pingErr := sqlDB.Ping()
	if pingErr != nil {
		panic(pingErr.Error())
	}

	fmt.Println("successfully connected to the database")
}
