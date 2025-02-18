package errors

import "errors"

var (
	ErrPolishWordNotFound       = errors.New("polish word not found")
	ErrEnglishWordNotFound      = errors.New("english word not found")
	ErrTranslationNotFound      = errors.New("translation not found")
	ErrExampleNotFound          = errors.New("example not found")
	ErrExampleAlreadyExists     = errors.New("example with this text already exists")
	ErrPolishWordAlreadyExists  = errors.New("polish word with this text already exists")
	ErrEnglishWordAlreadyExists = errors.New("english word with this text already exists")
)
