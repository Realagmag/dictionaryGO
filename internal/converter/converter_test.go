package converter

import (
	"testing"

	dbModels "github.com/realagmag/dictionaryGO/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestPolishToGraphType(t *testing.T) {
	converter := Converter{}
	polishWord := &dbModels.PolishWord{
		ID:   1,
		Text: "kot",
	}

	result := converter.PolishToGraphType(polishWord)

	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "kot", result.Text)
}

func TestEnglishToGraphType(t *testing.T) {
	converter := Converter{}

	englishWord := &dbModels.EnglishWord{
		ID:   2,
		Text: "cat",
	}

	result := converter.EnglishToGraphType(englishWord)

	assert.NotNil(t, result)
	assert.Equal(t, 2, result.ID)
	assert.Equal(t, "cat", result.Text)
}

func TestPolishSliceToGraphType(t *testing.T) {
	converter := Converter{}

	polishWords := []*dbModels.PolishWord{
		{ID: 1, Text: "kot"},
		{ID: 2, Text: "pies"},
	}

	result := converter.PolishSliceToGraphType(polishWords)

	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].ID)
	assert.Equal(t, "kot", result[0].Text)
	assert.Equal(t, 2, result[1].ID)
	assert.Equal(t, "pies", result[1].Text)
}

func TestEnglishSliceToGraphType(t *testing.T) {
	converter := Converter{}

	englishWords := []*dbModels.EnglishWord{
		{ID: 1, Text: "cat"},
		{ID: 2, Text: "dog"},
	}

	result := converter.EnglishSliceToGraphType(englishWords)

	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].ID)
	assert.Equal(t, "cat", result[0].Text)
	assert.Equal(t, 2, result[1].ID)
	assert.Equal(t, "dog", result[1].Text)
}

func TestTranslationToGraphType(t *testing.T) {
	converter := Converter{}

	translation := &dbModels.Translation{
		ID: 1,
		PolishWord: dbModels.PolishWord{
			ID:   1,
			Text: "kot",
		},
		EnglishWord: dbModels.EnglishWord{
			ID:   1,
			Text: "cat",
		},
		Examples: []dbModels.Example{
			{ID: 1, Text: "Kot jest zwierzęciem.", InPolish: true, TranslationID: 1},
		},
	}

	result := converter.TranslationToGraphType(translation)

	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "kot", result.PolishWord.Text)
	assert.Equal(t, "cat", result.EnglishWord.Text)
	assert.Len(t, result.Examples, 1)
	assert.Equal(t, "Kot jest zwierzęciem.", result.Examples[0].Text)
}

func TestExampleToGraphType(t *testing.T) {
	converter := Converter{}

	example := &dbModels.Example{
		ID:            1,
		Text:          "Kot jest zwierzęciem.",
		InPolish:      true,
		TranslationID: 1,
	}

	result := converter.ExampleToGraphType(example)

	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "Kot jest zwierzęciem.", result.Text)
	assert.True(t, result.InPolish)
	assert.Equal(t, 1, result.TranslationID)
}

func TestExampleSliceToGraphType(t *testing.T) {
	converter := Converter{}

	examples := []dbModels.Example{
		{ID: 1, Text: "Kot jest zwierzęciem.", InPolish: true, TranslationID: 1},
		{ID: 2, Text: "The cat is an animal.", InPolish: false, TranslationID: 1},
	}

	result := converter.ExampleSliceToGraphType(&examples)

	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].ID)
	assert.Equal(t, "Kot jest zwierzęciem.", result[0].Text)
	assert.Equal(t, 2, result[1].ID)
	assert.Equal(t, "The cat is an animal.", result[1].Text)
}
