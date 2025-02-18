package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	dbModels "github.com/realagmag/dictionaryGO/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_TIMEZONE"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
	fmt.Println("Connected to PostgreSQL!")

	migErr := DB.AutoMigrate(&dbModels.PolishWord{}, &dbModels.EnglishWord{}, &dbModels.Translation{}, &dbModels.Example{})
	if migErr != nil {
		log.Fatal("Failed to migrate database:", migErr)
	}
	// GORM didn't apply on delete cascade to foreign key
	db.Exec(`ALTER TABLE examples 
         DROP CONSTRAINT IF EXISTS fk_translations_examples;
         ALTER TABLE examples
         ADD CONSTRAINT fk_translations_examples 
         FOREIGN KEY (translation_id) 
         REFERENCES translations(id) 
         ON DELETE CASCADE;`)
	fmt.Println("Database migration complete!")
}
