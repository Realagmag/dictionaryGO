package database

import (
	"errors"
	"fmt"
	"strings"

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
	var polishWord dbModels.PolishWord
	err := manager.db.Where("text = ?", word).FirstOrCreate(&polishWord, dbModels.PolishWord{Text: word}).Error
	if err != nil {
		return nil, err
	}
	return &polishWord, nil
}

func (manager *DBManager) AddEnglishWord(word string) (*dbModels.EnglishWord, error) {
	var englishWord dbModels.EnglishWord
	err := manager.db.Where("text = ?", word).FirstOrCreate(&englishWord, dbModels.EnglishWord{Text: word}).Error
	if err != nil {
		return nil, err
	}
	return &englishWord, nil
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
	var translation dbModels.Translation
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

		translation = dbModels.Translation{
			PolishWordID:  polishWordModel.ID,
			EnglishWordID: englishWordModel.ID,
		}
		err = tx.Where("polish_word_id = ? AND english_word_id = ?", polishWordModel.ID, englishWordModel.ID).
			FirstOrCreate(&translation).Error
		if err != nil {
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
	return &translation, err
}

func (manager *DBManager) AddExampleToTranslation(example *model.ExampleInput, translationID uint) (*dbModels.Example, error) {
	var dbExample dbModels.Example
	err := manager.db.Where("translation_id = ? AND text = ?", translationID, example.Text).
		FirstOrCreate(&dbExample, dbModels.Example{
			TranslationID: translationID,
			Text:          example.Text,
			InPolish:      example.InPolish,
		}).Error

	if err != nil {
		return nil, err
	}
	return &dbExample, nil
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

func (manager *DBManager) DeleteRecordFromTable(table interface{}, id uint) error {
	return manager.db.Delete(&table, id).Error
}

func (manager *DBManager) ChangeExampleText(id uint, text string) (*dbModels.Example, error) {
	var example dbModels.Example
	err := manager.db.First(&example, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no Example record with id: %d", id)
		}
		return nil, err
	}
	example.Text = text
	if err := manager.db.Save(&example).Error; err != nil {
		if strings.Contains(err.Error(), "idx_translation_text") {
			return nil, fmt.Errorf("example with this text already exists")
		}
		return nil, err
	}
	return &example, nil
}

func (manager *DBManager) ChangePolishWordText(id uint, text string) (*dbModels.PolishWord, error) {
	var polishWord dbModels.PolishWord
	err := manager.db.First(&polishWord, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no PolishWord record with id: %d", id)
		}
		return nil, err
	}
	polishWord.Text = text
	if err := manager.db.Save(&polishWord).Error; err != nil {
		if strings.Contains(err.Error(), "uni_polish_words_text") {
			return nil, fmt.Errorf("polish word with this text already exists")
		}
		return nil, err
	}
	return &polishWord, nil
}
func (manager *DBManager) ChangeEnglishWordText(id uint, text string) (*dbModels.EnglishWord, error) {
	var englishWord dbModels.EnglishWord
	err := manager.db.First(&englishWord, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no EnglishWord record with id: %d", id)
		}
		return nil, err
	}
	englishWord.Text = text
	if err := manager.db.Save(&englishWord).Error; err != nil {
		if strings.Contains(err.Error(), "uni_english_words_text") {
			return nil, fmt.Errorf("english word with this text already exists")
		}
		return nil, err
	}
	return &englishWord, nil
}
