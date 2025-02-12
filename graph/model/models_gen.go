// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type EnglishWord struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type Example struct {
	ID       int    `json:"Id"`
	Text     string `json:"text"`
	InPolish *bool  `json:"inPolish,omitempty"`
}

type ExampleInput struct {
	Text     string `json:"text"`
	InPolish bool   `json:"inPolish"`
}

type IndividualExampleInput struct {
	TranslationID int           `json:"translationID"`
	Example       *ExampleInput `json:"example"`
}

type Mutation struct {
}

type PolishWord struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type Query struct {
}

type Translation struct {
	ID          int          `json:"Id"`
	PolishWord  *PolishWord  `json:"polishWord"`
	EnglishWord *EnglishWord `json:"englishWord"`
	Examples    []*Example   `json:"examples"`
}

type TranslationInput struct {
	PolishWord  string          `json:"polishWord"`
	EnglishWord string          `json:"englishWord"`
	Examples    []*ExampleInput `json:"examples,omitempty"`
}
