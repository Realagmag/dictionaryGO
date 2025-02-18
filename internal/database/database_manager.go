package database

import (
	"errors"
	"strings"

	"github.com/realagmag/dictionaryGO/graph/model"
	customErrors "github.com/realagmag/dictionaryGO/internal/errors"
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

	err := manager.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(`SELECT * FROM polish_words WHERE text = ? FOR UPDATE`, word).Scan(&polishWord).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if polishWord.ID == 0 {
			polishWord = dbModels.PolishWord{Text: word}
			if err := tx.Create(&polishWord).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "uni_polish_words_text") {
			return nil, customErrors.ErrPolishWordAlreadyExists
		}
		return nil, err
	}
	return &polishWord, nil
}

func (manager *DBManager) AddEnglishWord(word string) (*dbModels.EnglishWord, error) {
	var englishWord dbModels.EnglishWord

	err := manager.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(`SELECT * FROM english_words WHERE text = ? FOR UPDATE`, word).Scan(&englishWord).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if englishWord.ID == 0 {
			englishWord = dbModels.EnglishWord{Text: word}
			if err := tx.Create(&englishWord).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "uni_english_words_text") {
			return nil, customErrors.ErrEnglishWordAlreadyExists
		}
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
		for attempt := 0; attempt < 3; attempt++ {
			txManager := &DBManager{db: tx}

			polishWordModel, err := txManager.AddPolishWord(polishWord)
			if err != nil {
				if err == customErrors.ErrPolishWordAlreadyExists {
					continue
				}
				return err
			}
			englishWordModel, err := txManager.AddEnglishWord(englishWord)
			if err != nil {
				if err == customErrors.ErrEnglishWordAlreadyExists {
					continue
				}
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
				if _, err = txManager.AddExampleToTranslation(example, translation.ID); err != nil {
					return err
				}
			}

			return nil
		}
		return errors.New("failed to create translation after multiple retries")
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
		if strings.Contains(err.Error(), "fk_translations_examples") {
			return nil, customErrors.ErrTranslationNotFound
		}
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
			return nil, customErrors.ErrExampleNotFound
		}
		return nil, err
	}
	example.Text = text
	if err := manager.db.Save(&example).Error; err != nil {
		if strings.Contains(err.Error(), "idx_translation_text") {
			return nil, customErrors.ErrExampleAlreadyExists
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
			return nil, customErrors.ErrPolishWordNotFound
		}
		return nil, err
	}
	polishWord.Text = text
	if err := manager.db.Save(&polishWord).Error; err != nil {
		if strings.Contains(err.Error(), "uni_polish_words_text") {
			return nil, customErrors.ErrPolishWordAlreadyExists
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
			return nil, customErrors.ErrEnglishWordNotFound
		}
		return nil, err
	}
	englishWord.Text = text
	if err := manager.db.Save(&englishWord).Error; err != nil {
		if strings.Contains(err.Error(), "uni_english_words_text") {
			return nil, customErrors.ErrEnglishWordAlreadyExists
		}
		return nil, err
	}
	return &englishWord, nil
}

func (manager *DBManager) GetPolishWordById(id uint) (*dbModels.PolishWord, error) {
	var polishWord dbModels.PolishWord

	if err := manager.db.First(&polishWord, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrPolishWordNotFound
		}
		return nil, err
	}
	return &polishWord, nil
}

func (manager *DBManager) GetEnglishWordById(id uint) (*dbModels.EnglishWord, error) {
	var englishWord dbModels.EnglishWord

	if err := manager.db.First(&englishWord, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrEnglishWordNotFound
		}
		return nil, err
	}
	return &englishWord, nil
}

func (manager *DBManager) GetExampleById(id uint) (*dbModels.Example, error) {
	var example dbModels.Example

	if err := manager.db.First(&example, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrExampleNotFound
		}
		return nil, err
	}
	return &example, nil
}

func (manager *DBManager) GetTranslationById(id uint) (*dbModels.Translation, error) {
	var translation dbModels.Translation

	if err := manager.db.First(&translation, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrTranslationNotFound
		}
		return nil, err
	}
	return &translation, nil
}
