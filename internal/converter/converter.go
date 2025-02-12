package converter

import (
	"github.com/realagmag/dictionaryGO/graph/model"
	dbModels "github.com/realagmag/dictionaryGO/internal/models"
)

type Converter struct {
}

func (c *Converter) PolishToGraphType(word *dbModels.PolishWord) *model.PolishWord {
	return &model.PolishWord{
		ID:   int(word.ID),
		Text: word.Text,
	}
}

func (c *Converter) EnglishToGraphType(word *dbModels.EnglishWord) *model.EnglishWord {
	return &model.EnglishWord{
		ID:   int(word.ID),
		Text: word.Text,
	}
}

func (c *Converter) PolishSliceToGraphType(words []*dbModels.PolishWord) []*model.PolishWord {
	convertedWords := make([]*model.PolishWord, len(words))
	for i, word := range words {
		convertedWords[i] = c.PolishToGraphType(word)
	}
	return convertedWords
}

func (c *Converter) EnglishSliceToGraphType(words []*dbModels.EnglishWord) []*model.EnglishWord {
	convertedWords := make([]*model.EnglishWord, len(words))
	for i, word := range words {
		convertedWords[i] = c.EnglishToGraphType(word)
	}
	return convertedWords
}

func (c *Converter) TranslationToGraphType(translation *dbModels.Translation) *model.Translation {
	return &model.Translation{
		PolishWord:  c.PolishToGraphType(&translation.PolishWord),
		EnglishWord: c.EnglishToGraphType(&translation.EnglishWord),
		Examples:    c.ExampleSliceToGraphType(&translation.Examples),
	}
}

func (c *Converter) ExampleToGraphType(example *dbModels.Example) *model.Example {
	return &model.Example{
		ExampleID: int(example.ID),
		Text:      example.Text,
		InPolish:  &example.InPolish,
	}
}

func (c *Converter) ExampleSliceToGraphType(examples *[]dbModels.Example) []*model.Example {
	convertedExamples := make([]*model.Example, len(*examples))
	for i, example := range *examples {
		convertedExamples[i] = c.ExampleToGraphType(&example)
	}
	return convertedExamples
}
