package config

import (
	"fmt"
	"log"

	"github.com/realagmag/dictionaryGO/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := "host=localhost user=konrad password=konrad dbname=dictionary port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
	fmt.Println("Connected to PostgreSQL!")

	migErr := DB.AutoMigrate(&models.PolishWord{}, &models.EnglishWord{}, &models.Translation{}, &models.Example{})
	if migErr != nil {
		log.Fatal("Failed to migrate database:", migErr)
	}

	fmt.Println("Database migration complete!")
}
