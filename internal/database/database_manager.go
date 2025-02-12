package database

import (
	"github.com/realagmag/dictionaryGO/graph/model"
	dbModels "github.com/realagmag/dictionaryGO/internal/models"
	"gorm.io/gorm"
)

type DBManager struct {
	db *gorm.DB
}

func NewDBManager(db *gorm.DB) *DBManager {
	return &DBManager{db: db}
}

func (manager *DBManager) AddPolishWord(word string) (*dbModels.PolishWord, error) {
	var existingWord dbModels.PolishWord
	if err := manager.db.Where("text = ?", word).First(&existingWord).Error; err == nil {
		return &existingWord, nil
	}

	newWord := &dbModels.PolishWord{
		Text: word,
	}
	if err := manager.db.Create(newWord).Error; err != nil {
		return nil, err
	}
	return newWord, nil
}

func (manager *DBManager) AddEnglishWord(word string) (*dbModels.EnglishWord, error) {
	var existingWord dbModels.EnglishWord
	if err := manager.db.Where("text = ?", word).First(&existingWord).Error; err == nil {
		return &existingWord, nil
	}

	newWord := &dbModels.EnglishWord{
		Text: word,
	}
	if err := manager.db.Create(newWord).Error; err != nil {
		return nil, err
	}
	return newWord, nil
}

func (manager *DBManager) GetPolishWords() ([]*dbModels.PolishWord, error) {
	var words []*dbModels.PolishWord
	if err := manager.db.Find(&words).Error; err != nil {
		return nil, err
	}
	return words, nil
}

func (manager *DBManager) GetEnglishWords() ([]*dbModels.EnglishWord, error) {
	var words []*dbModels.EnglishWord
	if err := manager.db.Find(&words).Error; err != nil {
		return nil, err
	}
	return words, nil
}

func (manager *DBManager) AddTranslation(translationInput model.TranslationInput) (*dbModels.Translation, error) {
	polishWord := translationInput.PolishWord
	englishWord := translationInput.EnglishWord
	examples := translationInput.Examples
	var translation *dbModels.Translation
	err := manager.db.Transaction(func(tx *gorm.DB) error {
		originalDB := manager.db
		manager.db = tx
		defer func() { manager.db = originalDB }()
		polishWordModel, err := manager.AddPolishWord(polishWord)
		if err != nil {
			return err
		}
		englishWordModel, err := manager.AddEnglishWord(englishWord)
		if err != nil {
			return err
		}

		translation = &dbModels.Translation{
			PolishWordID:  polishWordModel.ID,
			EnglishWordID: englishWordModel.ID,
			PolishWord:    *polishWordModel,
			EnglishWord:   *englishWordModel,
		}
		if err := tx.Create(translation).Error; err != nil {
			return err
		}

		for _, example := range examples {
			if _, err = manager.AddExampleToTranslation(example, translation.ID); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return translation, err
}

func (manager *DBManager) AddExampleToTranslation(example *model.ExampleInput, translationID uint) (*dbModels.Example, error) {
	newExample := &dbModels.Example{
		TranslationID: translationID,
		Text:          example.Text,
		InPolish:      example.InPolish,
	}
	if err := manager.db.Create(newExample).Error; err != nil {
		return nil, err
	}
	return newExample, nil
}

func (manager *DBManager) PopulateTranslationWithAssociations(translation *dbModels.Translation) error {
	return manager.db.Preload("PolishWord").Preload("EnglishWord").Preload("Examples").First(translation).Error
}

func (manager *DBManager) GetTranslations() ([]*dbModels.Translation, error) {
	var translations []*dbModels.Translation
	if err := manager.db.Find(&translations).Error; err != nil {
		return nil, err
	}
	return translations, nil
}

func (manager *DBManager) GetTranslationsToEnglish(wordInPolish string) ([]*dbModels.Translation, error) {
	var translations []*dbModels.Translation
	if err := manager.db.
		Where("polish_word_id IN (SELECT id FROM polish_words WHERE text = ?)", wordInPolish).
		Find(&translations).Error; err != nil {
		return nil, err
	}
	return translations, nil
}

func (manager *DBManager) GetTranslationsToPolish(wordInEnglish string) ([]*dbModels.Translation, error) {
	var translations []*dbModels.Translation
	if err := manager.db.
		Where("english_word_id IN (SELECT id FROM english_words WHERE text = ?)", wordInEnglish).
		Find(&translations).Error; err != nil {
		return nil, err
	}
	return translations, nil
}
