package customerrors

import "errors"

var (
	ErrDataNotFound    = errors.New("data not found")
	ErrInvalidUserCode = errors.New("invalid user code")
	ErrInvalidPassword = errors.New("invalid password")
	ErrFailedSaveToken = errors.New("fail save token")
	ErrNoDataEdited = errors.New("no data edited")
	ErrDataAlreadyExists = errors.New("data already exists")
	ErrInvalidClaims = errors.New("invalid claims")
	ErrInvalidToken = errors.New("invalid token or claims")
	ErrKeyNotFound = errors.New("key not found")
	ErrInvalidArrayFormat = errors.New("invalid array format")
	ErrInvalidQuantity = errors.New("invalid quantity")
	ErrInvalidInput = errors.New("invalid input")
)
