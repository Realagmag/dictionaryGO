package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"github.com/realagmag/dictionaryGO/graph/model"
	"github.com/realagmag/dictionaryGO/internal/converter"
	"github.com/realagmag/dictionaryGO/internal/database"
	dbModels "github.com/realagmag/dictionaryGO/internal/models"
)

type Resolver struct {
	DBManager *database.DBManager
	Converter *converter.Converter
}

func (r *Resolver) PrepareTranslationSliceToSend(translationDbModels *[]*dbModels.Translation) ([]*model.Translation, error) {
	translations := make([]*model.Translation, len(*translationDbModels))
	for i, translationDbModel := range *translationDbModels {
		if err := r.DBManager.PopulateTranslationWithAssociations(translationDbModel); err != nil {
			return nil, err
		}
		translations[i] = r.Converter.TranslationToGraphType(translationDbModel)
	}
	return translations, nil
}
