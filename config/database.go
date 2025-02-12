package config

import (
	"fmt"
	"log"

	dbModels "github.com/realagmag/dictionaryGO/internal/models"

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
