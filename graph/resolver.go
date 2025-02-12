package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"github.com/realagmag/dictionaryGO/internal/converter"
	"github.com/realagmag/dictionaryGO/internal/database"
)

type Resolver struct {
	DBManager *database.DBManager
	Converter *converter.Converter
}
