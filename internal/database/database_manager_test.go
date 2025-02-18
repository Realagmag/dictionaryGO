package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/joho/godotenv"
	"github.com/realagmag/dictionaryGO/graph/model"
	customErrors "github.com/realagmag/dictionaryGO/internal/errors"
	dbModels "github.com/realagmag/dictionaryGO/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var manager *DBManager

func setupTestDB() {
	projectRoot, err := filepath.Abs("../../")
	if err != nil {
		log.Fatal("Failed to get project root path:", err)
	}
	envPath := filepath.Join(projectRoot, ".env")
	err = godotenv.Load(envPath)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_TIMEZONE"),
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to the database: %v", err)
	}

	migErr := db.AutoMigrate(&dbModels.PolishWord{}, &dbModels.EnglishWord{}, &dbModels.Translation{}, &dbModels.Example{})
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
	manager = NewDBManager(db)
}

func clearTestDB(db *gorm.DB) {
	db.Exec("TRUNCATE TABLE examples, translations, english_words, polish_words RESTART IDENTITY CASCADE;")
}

func TestMain(m *testing.M) {
	setupTestDB()

	code := m.Run()
	os.Exit(code)
}

func TestAddPolishWord(t *testing.T) {
	defer clearTestDB(manager.db)

	word := "koń"

	polishWord, err := manager.AddPolishWord(word)
	assert.NoError(t, err)
	assert.Equal(t, word, polishWord.Text)
}

func TestAddSamePolishWord(t *testing.T) {
	defer clearTestDB(manager.db)

	word := "kot"

	polishWord, err := manager.AddPolishWord(word)
	assert.NoError(t, err)
	assert.Equal(t, word, polishWord.Text)

	polishWord2, err := manager.AddPolishWord(word)
	assert.NoError(t, err)
	assert.Equal(t, polishWord.ID, polishWord2.ID)
	assert.Equal(t, polishWord.Text, polishWord2.Text)
}

func TestAddDifferentPolishWord(t *testing.T) {
	defer clearTestDB(manager.db)

	word := "kot"

	polishWord, err := manager.AddPolishWord(word)
	assert.NoError(t, err)
	assert.Equal(t, word, polishWord.Text)

	word2 := "pies"
	polishWord2, err := manager.AddPolishWord(word2)
	assert.NoError(t, err)
	assert.NotEqual(t, polishWord.ID, polishWord2.ID)
	assert.NotEqual(t, polishWord.Text, polishWord2.Text)
}

func TestAddEnglishWord(t *testing.T) {
	defer clearTestDB(manager.db)

	word := "cat"

	englishWord, err := manager.AddEnglishWord(word)
	assert.NoError(t, err)
	assert.Equal(t, word, englishWord.Text)
}

func TestAddSameEnglishWord(t *testing.T) {
	defer clearTestDB(manager.db)

	word := "cat"

	englishWord, err := manager.AddEnglishWord(word)
	assert.NoError(t, err)
	assert.Equal(t, word, englishWord.Text)

	englishWord2, err := manager.AddEnglishWord(word)
	assert.NoError(t, err)
	assert.Equal(t, englishWord.ID, englishWord2.ID)
}

func TestCreateTranslationNewWordsNoExamples(t *testing.T) {
	defer clearTestDB(manager.db)
	translation, err := manager.AddTranslation(model.TranslationInput{PolishWord: "kot", EnglishWord: "cat"})
	assert.NoError(t, err)
	assert.NotNil(t, translation)
}

func TestPopulateTranslationWithAssociacions(t *testing.T) {
	defer clearTestDB(manager.db)

	plWord := "kot"
	enWord := "cat"
	translation, err := manager.AddTranslation(model.TranslationInput{PolishWord: plWord, EnglishWord: enWord})
	assert.NoError(t, err)
	assert.Empty(t, translation.PolishWord.Text)
	assert.Empty(t, translation.EnglishWord.Text)
	assert.Empty(t, translation.Examples)
	err = manager.PopulateTranslationWithAssociations(translation)
	assert.NoError(t, err)
	assert.Equal(t, plWord, translation.PolishWord.Text)
	assert.Equal(t, enWord, translation.EnglishWord.Text)
	assert.Empty(t, translation.Examples)
}

func TestAddTranslationExistingWordsNoExamples(t *testing.T) {
	defer clearTestDB(manager.db)

	enWord := "cat"
	englishWord, err := manager.AddEnglishWord(enWord)
	assert.NoError(t, err)
	plWord := "kot"
	polishWord, err := manager.AddPolishWord(plWord)
	assert.NoError(t, err)
	translationInput := model.TranslationInput{PolishWord: plWord, EnglishWord: enWord}
	translation, err := manager.AddTranslation(translationInput)
	assert.NoError(t, err)
	err = manager.PopulateTranslationWithAssociations(translation)
	assert.NoError(t, err)
	assert.Equal(t, translation.PolishWord.Text, polishWord.Text)
	assert.Equal(t, translation.PolishWord.ID, polishWord.ID)
	assert.Equal(t, translation.EnglishWord.Text, englishWord.Text)
	assert.Equal(t, translation.EnglishWord.ID, englishWord.ID)
	assert.Empty(t, translation.Examples)
}

func TestAddTranslationNewWordsWithExamples(t *testing.T) {
	defer clearTestDB(manager.db)

	plWord := "kot"
	enWord := "cat"
	translation, err := manager.AddTranslation(
		model.TranslationInput{
			PolishWord:  plWord,
			EnglishWord: enWord,
			Examples: []*model.ExampleInput{
				{Text: "Ala ma kota", InPolish: true},
				{Text: "Kot ma Alę", InPolish: true},
				{Text: "Cats sleep all day long.", InPolish: false},
			}})
	assert.NoError(t, err)
	err = manager.PopulateTranslationWithAssociations(translation)
	assert.NoError(t, err)
	assert.Equal(t, plWord, translation.PolishWord.Text)
	assert.Equal(t, enWord, translation.EnglishWord.Text)
	assert.Len(t, translation.Examples, 3)
	assert.Equal(t, translation.Examples[0].Text, "Ala ma kota")
	assert.Equal(t, translation.Examples[1].Text, "Kot ma Alę")
	assert.Equal(t, translation.Examples[2].Text, "Cats sleep all day long.")
	assert.Equal(t, translation.Examples[0].InPolish, true)
	assert.Equal(t, translation.Examples[1].InPolish, true)
	assert.Equal(t, translation.Examples[2].InPolish, false)
}

func TestGetTranslations(t *testing.T) {
	defer clearTestDB(manager.db)

	manager.AddTranslation(
		model.TranslationInput{
			PolishWord:  "kot",
			EnglishWord: "cat",
			Examples: []*model.ExampleInput{
				{Text: "Ala ma kota", InPolish: true},
			}})
	translations, err := manager.GetTranslations()
	assert.NoError(t, err)
	assert.Len(t, translations, 1)
	translation := translations[0]
	err = manager.PopulateTranslationWithAssociations(translation)
	assert.NoError(t, err)
	assert.Equal(t, "kot", translation.PolishWord.Text)
	assert.Equal(t, "cat", translation.EnglishWord.Text)
	assert.Equal(t, "kot", translation.PolishWord.Text)
	assert.Equal(t, "Ala ma kota", translation.Examples[0].Text)
	assert.Equal(t, true, translation.Examples[0].InPolish)
}

func TestAddExampleToExistingTranslation(t *testing.T) {
	defer clearTestDB(manager.db)

	plWord := "kot"
	enWord := "cat"
	translation, err := manager.AddTranslation(
		model.TranslationInput{
			PolishWord:  plWord,
			EnglishWord: enWord,
			Examples: []*model.ExampleInput{
				{Text: "Ala ma kota", InPolish: true},
				{Text: "Kot ma Alę", InPolish: true},
				{Text: "Cats sleep all day long.", InPolish: false},
			}})
	assert.NoError(t, err)
	Example, err := manager.AddExampleToTranslation(
		&model.ExampleInput{Text: "Cats are cute.", InPolish: false},
		uint(translation.ID),
	)
	assert.NoError(t, err)
	assert.Equal(t, "Cats are cute.", Example.Text)
	assert.Equal(t, false, Example.InPolish)
	assert.Equal(t, translation.ID, Example.TranslationID)
	manager.PopulateTranslationWithAssociations(translation)
	assert.Len(t, translation.Examples, 4)
}

func TestAddSimilarExampleToExistingTranslation(t *testing.T) {
	defer clearTestDB(manager.db)

	plWord := "kot"
	enWord := "cat"
	translation, err := manager.AddTranslation(
		model.TranslationInput{
			PolishWord:  plWord,
			EnglishWord: enWord,
			Examples: []*model.ExampleInput{
				{Text: "Ala ma kota", InPolish: true},
				{Text: "Kot ma Alę", InPolish: true},
				{Text: "Cats sleep all day long.", InPolish: false},
			}})
	assert.NoError(t, err)
	newExample := &model.ExampleInput{Text: "Cats are cute.", InPolish: false}
	Example, err := manager.AddExampleToTranslation(
		newExample,
		uint(translation.ID),
	)
	assert.NoError(t, err)
	assert.Equal(t, "Cats are cute.", Example.Text)
	assert.Equal(t, false, Example.InPolish)
	assert.Equal(t, translation.ID, Example.TranslationID)
	manager.PopulateTranslationWithAssociations(translation)
	assert.Len(t, translation.Examples, 4)
	Example, err = manager.AddExampleToTranslation(
		newExample,
		uint(translation.ID),
	)
	assert.NoError(t, err)
	assert.Equal(t, "Cats are cute.", Example.Text)
	assert.Equal(t, false, Example.InPolish)
	assert.Equal(t, translation.ID, Example.TranslationID)
	manager.PopulateTranslationWithAssociations(translation)
	assert.Len(t, translation.Examples, 4)
}

func TestGetPolishWords(t *testing.T) {
	defer clearTestDB(manager.db)

	manager.AddPolishWord("kotlet")
	manager.AddPolishWord("świeca")
	manager.AddPolishWord("kaczka")
	polishWords, err := manager.GetPolishWords()
	assert.NoError(t, err)
	assert.Len(t, polishWords, 3)
	assert.Equal(t, "kotlet", polishWords[0].Text)
	assert.Equal(t, "świeca", polishWords[1].Text)
	assert.Equal(t, "kaczka", polishWords[2].Text)
}

func TestGetEnglishWords(t *testing.T) {
	defer clearTestDB(manager.db)

	manager.AddEnglishWord("meal")
	manager.AddEnglishWord("dream")
	englishWords, err := manager.GetEnglishWords()
	assert.NoError(t, err)
	assert.Len(t, englishWords, 2)
	assert.Equal(t, "meal", englishWords[0].Text)
	assert.Equal(t, "dream", englishWords[1].Text)
}

func TestGetTranslationsToEnglish(t *testing.T) {
	defer clearTestDB(manager.db)
	manager.AddTranslation(model.TranslationInput{
		PolishWord:  "wieża",
		EnglishWord: "tower",
	})
	manager.AddTranslation(model.TranslationInput{
		PolishWord:  "wieża",
		EnglishWord: "rook",
	})
	manager.AddTranslation(model.TranslationInput{
		PolishWord:  "koń",
		EnglishWord: "horse",
	})
	translationsToEnglish, err := manager.GetTranslationsToEnglish("wieża")
	assert.NoError(t, err)
	assert.Len(t, translationsToEnglish, 2)
	firstTranslation := translationsToEnglish[0]
	manager.PopulateTranslationWithAssociations(firstTranslation)
	assert.Equal(t, "tower", firstTranslation.EnglishWord.Text)
	assert.Equal(t, "wieża", firstTranslation.PolishWord.Text)
	secondTranslation := translationsToEnglish[1]
	manager.PopulateTranslationWithAssociations(secondTranslation)
	assert.Equal(t, "rook", secondTranslation.EnglishWord.Text)
	assert.Equal(t, "wieża", secondTranslation.PolishWord.Text)

	translationsToEnglish, err = manager.GetTranslationsToEnglish("koń")
	assert.NoError(t, err)
	assert.Len(t, translationsToEnglish, 1)
	manager.PopulateTranslationWithAssociations(translationsToEnglish[0])
	assert.Equal(t, "horse", translationsToEnglish[0].EnglishWord.Text)
	assert.Equal(t, "koń", translationsToEnglish[0].PolishWord.Text)
}

func TestGetTranslationsToPolish(t *testing.T) {
	defer clearTestDB(manager.db)
	manager.AddTranslation(model.TranslationInput{
		PolishWord:  "zarezerwować",
		EnglishWord: "book",
	})
	manager.AddTranslation(model.TranslationInput{
		PolishWord:  "książka",
		EnglishWord: "book",
	})
	manager.AddTranslation(model.TranslationInput{
		PolishWord:  "koń",
		EnglishWord: "horse",
	})
	translationsToPolish, err := manager.GetTranslationsToPolish("book")
	assert.NoError(t, err)
	assert.Len(t, translationsToPolish, 2)
	firstTranslation := translationsToPolish[0]
	manager.PopulateTranslationWithAssociations(firstTranslation)
	assert.Equal(t, "book", firstTranslation.EnglishWord.Text)
	assert.Equal(t, "zarezerwować", firstTranslation.PolishWord.Text)
	secondTranslation := translationsToPolish[1]
	manager.PopulateTranslationWithAssociations(secondTranslation)
	assert.Equal(t, "book", secondTranslation.EnglishWord.Text)
	assert.Equal(t, "książka", secondTranslation.PolishWord.Text)

	translationsToPolish, err = manager.GetTranslationsToEnglish("koń")
	assert.NoError(t, err)
	assert.Len(t, translationsToPolish, 1)
	manager.PopulateTranslationWithAssociations(translationsToPolish[0])
	assert.Equal(t, "horse", translationsToPolish[0].EnglishWord.Text)
	assert.Equal(t, "koń", translationsToPolish[0].PolishWord.Text)
}

func TestDeletePolishWord(t *testing.T) {
	defer clearTestDB(manager.db)

	manager.AddPolishWord("praca")
	manager.AddPolishWord("stół")
	manager.AddPolishWord("książka")
	polishWords, _ := manager.GetPolishWords()
	err := manager.DeleteRecordFromTable(dbModels.PolishWord{}, polishWords[0].ID)
	assert.NoError(t, err)
	polishWords, err = manager.GetPolishWords()
	assert.NoError(t, err)
	assert.Len(t, polishWords, 2)
	assert.Equal(t, "stół", polishWords[0].Text)
	assert.Equal(t, "książka", polishWords[1].Text)
}

func TestDeleteEnglishWord(t *testing.T) {
	defer clearTestDB(manager.db)

	manager.AddEnglishWord("work")
	manager.AddEnglishWord("table")
	manager.AddEnglishWord("book")
	englishWords, _ := manager.GetEnglishWords()
	err := manager.DeleteRecordFromTable(dbModels.EnglishWord{}, englishWords[0].ID)
	assert.NoError(t, err)
	englishWords, err = manager.GetEnglishWords()
	assert.NoError(t, err)
	assert.Len(t, englishWords, 2)
	assert.Equal(t, "table", englishWords[0].Text)
	assert.Equal(t, "book", englishWords[1].Text)
}

func TestDeletePolishWordCascase(t *testing.T) {
	defer clearTestDB(manager.db)

	manager.AddTranslation(model.TranslationInput{PolishWord: "książka", EnglishWord: "book"})
	translations, _ := manager.GetTranslations()
	assert.Len(t, translations, 1)
	err := manager.DeleteRecordFromTable(dbModels.PolishWord{}, translations[0].PolishWordID)
	assert.NoError(t, err)
	translations, err = manager.GetTranslations()
	assert.NoError(t, err)
	assert.Len(t, translations, 0)
	polishWords, _ := manager.GetPolishWords()
	assert.Len(t, polishWords, 0)
	englishWords, _ := manager.GetEnglishWords()
	assert.Len(t, englishWords, 1)
}

func TestDeleteEnglishWordCascase(t *testing.T) {
	defer clearTestDB(manager.db)

	manager.AddTranslation(model.TranslationInput{PolishWord: "książka", EnglishWord: "book"})
	translations, _ := manager.GetTranslations()
	assert.Len(t, translations, 1)
	err := manager.DeleteRecordFromTable(dbModels.EnglishWord{}, translations[0].EnglishWordID)
	assert.NoError(t, err)
	translations, err = manager.GetTranslations()
	assert.NoError(t, err)
	assert.Len(t, translations, 0)
	polishWords, _ := manager.GetPolishWords()
	assert.Len(t, polishWords, 1)
	englishWords, _ := manager.GetEnglishWords()
	assert.Len(t, englishWords, 0)
}

func TestDeleteExample(t *testing.T) {
	defer clearTestDB(manager.db)

	translation, _ := manager.AddTranslation(model.TranslationInput{
		PolishWord:  "książka",
		EnglishWord: "book",
		Examples: []*model.ExampleInput{
			{Text: "Książki mają strony", InPolish: true},
			{Text: "Books are heavy", InPolish: false},
			{Text: "Kupiłem ciekawą książkę", InPolish: true},
		},
	})
	manager.PopulateTranslationWithAssociations(translation)
	assert.Len(t, translation.Examples, 3)
	err := manager.DeleteRecordFromTable(dbModels.Example{}, translation.Examples[0].ID)
	assert.NoError(t, err)
	manager.PopulateTranslationWithAssociations(translation)
	assert.Len(t, translation.Examples, 2)
}

func TestDeleteTranslation(t *testing.T) {
	defer clearTestDB(manager.db)

	translation, _ := manager.AddTranslation(model.TranslationInput{
		PolishWord:  "dziecko",
		EnglishWord: "child",
	})
	err := manager.DeleteRecordFromTable(dbModels.Translation{}, translation.ID)
	assert.NoError(t, err)
	translations, _ := manager.GetTranslations()
	assert.Len(t, translations, 0)
	polishWords, _ := manager.GetPolishWords()
	assert.Len(t, polishWords, 1)
	assert.Equal(t, "dziecko", polishWords[0].Text)
	englishWords, _ := manager.GetEnglishWords()
	assert.Len(t, englishWords, 1)
	assert.Equal(t, "child", englishWords[0].Text)
}

func TestDeleteTranslationCascadeOnExamples(t *testing.T) {
	defer clearTestDB(manager.db)

	translation, _ := manager.AddTranslation(model.TranslationInput{
		PolishWord:  "dziecko",
		EnglishWord: "child",
		Examples: []*model.ExampleInput{
			{Text: "Dziecko je cukierka", InPolish: true},
			{Text: "Children are playing outside", InPolish: false},
		},
	})
	manager.PopulateTranslationWithAssociations(translation)
	examples := translation.Examples
	assert.Len(t, examples, 2)
	firstExampleID := examples[0].ID
	secondExampleID := examples[1].ID
	err := manager.DeleteRecordFromTable(dbModels.Translation{}, translation.ID)
	assert.NoError(t, err)
	firstExample, err := manager.GetExampleById(firstExampleID)
	assert.Equal(t, customErrors.ErrExampleNotFound, err)
	assert.Nil(t, firstExample)
	secondExample, err := manager.GetExampleById(secondExampleID)
	assert.Equal(t, customErrors.ErrExampleNotFound, err)
	assert.Nil(t, secondExample)
}

func TestChangeExampleText(t *testing.T) {
	defer clearTestDB(manager.db)

	translation, _ := manager.AddTranslation(model.TranslationInput{
		PolishWord:  "dziecko",
		EnglishWord: "child",
		Examples: []*model.ExampleInput{
			{Text: "Dziecko je cukierka", InPolish: true},
		},
	})
	manager.PopulateTranslationWithAssociations(translation)
	assert.Equal(t, "Dziecko je cukierka", translation.Examples[0].Text)
	newText := "Dziecko chodzi do przedszkola"
	example, err := manager.ChangeExampleText(translation.Examples[0].ID, newText)
	assert.NoError(t, err)
	assert.Equal(t, translation.Examples[0].ID, example.ID)
	assert.NotEqual(t, translation.Examples[0].Text, example.Text)
	translation, _ = manager.GetTranslationById(translation.ID)
	manager.PopulateTranslationWithAssociations(translation)
	assert.Equal(t, translation.Examples[0].Text, "Dziecko chodzi do przedszkola")
}

func TestChangeExampleTextViolatesUniqueConstraint(t *testing.T) {
	defer clearTestDB(manager.db)

	translation, _ := manager.AddTranslation(model.TranslationInput{
		PolishWord:  "dziecko",
		EnglishWord: "child",
		Examples: []*model.ExampleInput{
			{Text: "Dziecko je cukierka", InPolish: true},
		},
	})
	newExample := model.ExampleInput{Text: "Dziecko chodzi do przedszkola", InPolish: true}
	manager.AddExampleToTranslation(&newExample, translation.ID)
	manager.PopulateTranslationWithAssociations(translation)
	_, err := manager.ChangeExampleText(translation.Examples[1].ID, "Dziecko je cukierka")
	assert.Error(t, err)
	assert.Equal(t, customErrors.ErrExampleAlreadyExists, err)
}

func TestChangeExampleTextOnNonExistingExample(t *testing.T) {
	defer clearTestDB(manager.db)

	_, err := manager.ChangeExampleText(555, "Dziecko je cukierka")
	assert.Error(t, err)
	assert.Equal(t, customErrors.ErrExampleNotFound, err)
}

func TestPreventDuplicatePolishWordsWhileCreating(t *testing.T) {
	defer clearTestDB(manager.db)

	var wg sync.WaitGroup
	concurrency := 100

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = manager.AddPolishWord("balon")
		}()
	}

	wg.Wait()

	polishWords, err := manager.GetPolishWords()
	assert.Len(t, polishWords, 1, "Only one 'balon' record should exist in the database")
	assert.NoError(t, err)
}

func TestChangePolishWordTextCorrect(t *testing.T) {
	defer clearTestDB(manager.db)

	originalWord, _ := manager.AddPolishWord("książka")
	polishWords, _ := manager.GetPolishWords()
	word, err := manager.ChangePolishWordText(polishWords[0].ID, "stół")
	assert.NoError(t, err)
	assert.Equal(t, originalWord.ID, word.ID)
	assert.NotEqual(t, originalWord.Text, word.Text)
}

func TestPreventDuplicatePolishWordsWhileUpdating(t *testing.T) {
	defer clearTestDB(manager.db)

	manager.AddPolishWord("książka")
	manager.AddPolishWord("miecz")
	polishWords, _ := manager.GetPolishWords()
	assert.Len(t, polishWords, 2)
	word, err := manager.ChangePolishWordText(polishWords[0].ID, polishWords[1].Text)
	assert.Nil(t, word)
	assert.Equal(t, customErrors.ErrPolishWordAlreadyExists, err)
	polishWords, _ = manager.GetPolishWords()
	assert.Len(t, polishWords, 2)
}

func TestChangePolishWordOnNonExistingPolishWord(t *testing.T) {
	defer clearTestDB(manager.db)

	word, err := manager.ChangePolishWordText(555, "przykład")
	assert.Error(t, err)
	assert.Equal(t, customErrors.ErrPolishWordNotFound, err)
	assert.Nil(t, word)
}

func TestPreventDuplicateEnglishWordsWhileCreating(t *testing.T) {
	defer clearTestDB(manager.db)

	var wg sync.WaitGroup
	concurrency := 100

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = manager.AddEnglishWord("baloon")
		}()
	}

	wg.Wait()

	englishWords, err := manager.GetEnglishWords()
	assert.Len(t, englishWords, 1, "Only one 'baloon' record should exist in the database")
	assert.NoError(t, err)
}

func TestChangeEnglishWordTextCorrect(t *testing.T) {
	defer clearTestDB(manager.db)

	originalWord, _ := manager.AddEnglishWord("book")
	englishWords, _ := manager.GetEnglishWords()
	word, err := manager.ChangeEnglishWordText(englishWords[0].ID, "table")
	assert.NoError(t, err)
	assert.Equal(t, originalWord.ID, word.ID)
	assert.NotEqual(t, originalWord.Text, word.Text)
}

func TestPreventDuplicateEnglishWordsWhileUpdating(t *testing.T) {
	defer clearTestDB(manager.db)

	manager.AddEnglishWord("book")
	manager.AddEnglishWord("sword")
	englishWords, _ := manager.GetEnglishWords()
	assert.Len(t, englishWords, 2)
	word, err := manager.ChangeEnglishWordText(englishWords[0].ID, englishWords[1].Text)
	assert.Nil(t, word)
	assert.Equal(t, customErrors.ErrEnglishWordAlreadyExists, err)
	englishWords, _ = manager.GetEnglishWords()
	assert.Len(t, englishWords, 2)
}

func TestChangeEnglishWordOnNonExistingEnglishWord(t *testing.T) {
	defer clearTestDB(manager.db)

	word, err := manager.ChangeEnglishWordText(555, "example")
	assert.Error(t, err)
	assert.Equal(t, customErrors.ErrEnglishWordNotFound, err)
	assert.Nil(t, word)
}

func TestPreventDuplicateWordsWhileCreatingTranslations(t *testing.T) {
	defer clearTestDB(manager.db)

	var wg sync.WaitGroup
	concurrency := 100

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = manager.AddTranslation(model.TranslationInput{
				PolishWord:  "chleb",
				EnglishWord: "bread",
				Examples: []*model.ExampleInput{
					{Text: "test", InPolish: false},
				},
			})
		}()
	}

	wg.Wait()

	englishWords, err := manager.GetEnglishWords()
	assert.Len(t, englishWords, 1, "Only one record should exist in the database")
	assert.NoError(t, err)
	polishWords, err := manager.GetPolishWords()
	assert.Len(t, polishWords, 1, "Only one record should exist in the database")
	assert.NoError(t, err)
	translations, err := manager.GetTranslations()
	assert.Len(t, translations, 1, "Only one record should exist in the database")
	assert.NoError(t, err)
	translation := translations[0]
	manager.PopulateTranslationWithAssociations(translation)
	assert.Len(t, translation.Examples, 1, "Duplicate examples should not be added to database")
}

func TestAddEveryExampleToTranslationWhileConcurentRequests(t *testing.T) {
	defer clearTestDB(manager.db)

	var wg sync.WaitGroup
	concurrency := 100

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = manager.AddTranslation(model.TranslationInput{
				PolishWord:  "chleb",
				EnglishWord: "bread",
				Examples: []*model.ExampleInput{
					{Text: fmt.Sprintf("test %v", i), InPolish: false},
				},
			})
		}(i)
	}

	wg.Wait()

	englishWords, err := manager.GetEnglishWords()
	assert.Len(t, englishWords, 1, "Only one record should exist in the database")
	assert.NoError(t, err)
	polishWords, err := manager.GetPolishWords()
	assert.Len(t, polishWords, 1, "Only one record should exist in the database")
	assert.NoError(t, err)
	translations, err := manager.GetTranslations()
	assert.Len(t, translations, 1, "Only one record should exist in the database")
	assert.NoError(t, err)
	translation := translations[0]
	manager.PopulateTranslationWithAssociations(translation)
	for _, example := range translation.Examples {
		fmt.Println(example.ID, example.Text)
	}
	assert.Len(t, translation.Examples, 100, "Each thread adds its unique example to translation")
}

func TestGetPolishWordByIdCorrect(t *testing.T) {
	defer clearTestDB(manager.db)
	polishWord, _ := manager.AddPolishWord("słowo")
	word, err := manager.GetPolishWordById(polishWord.ID)
	assert.NoError(t, err)
	assert.Equal(t, polishWord, word)
}

func TestGetPolishWordByIdWordNotExists(t *testing.T) {
	defer clearTestDB(manager.db)
	word, err := manager.GetPolishWordById(555)
	assert.Error(t, err)
	assert.Equal(t, customErrors.ErrPolishWordNotFound, err)
	assert.Nil(t, word)
}

func TestGetEnglishWordByIdCorrect(t *testing.T) {
	defer clearTestDB(manager.db)
	englishWord, _ := manager.AddEnglishWord("word")
	word, err := manager.GetEnglishWordById(englishWord.ID)
	assert.NoError(t, err)
	assert.Equal(t, englishWord, word)
}

func TestGetEnglishWordByIdWordNotExists(t *testing.T) {
	defer clearTestDB(manager.db)
	word, err := manager.GetEnglishWordById(555)
	assert.Error(t, err)
	assert.Equal(t, customErrors.ErrEnglishWordNotFound, err)
	assert.Nil(t, word)
}

func TestGetExampleByIdCorrect(t *testing.T) {
	defer clearTestDB(manager.db)
	translation, _ := manager.AddTranslation(
		model.TranslationInput{
			PolishWord:  "poduszka",
			EnglishWord: "pillow",
			Examples: []*model.ExampleInput{
				{Text: "miękka poduszka", InPolish: true},
			},
		},
	)
	manager.PopulateTranslationWithAssociations(translation)
	example, err := manager.GetExampleById(translation.Examples[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, "miękka poduszka", example.Text)
}

func TestGetTranslationByIdTranslationNotExists(t *testing.T) {
	defer clearTestDB(manager.db)
	translation, err := manager.GetTranslationById(555)
	assert.Error(t, err)
	assert.Equal(t, customErrors.ErrTranslationNotFound, err)
	assert.Nil(t, translation)
}
