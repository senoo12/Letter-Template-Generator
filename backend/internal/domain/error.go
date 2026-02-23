package domain

import "errors"

var (
	ErrMissingColumns      = errors.New("excel missing required columns")
	ErrTemplateInvalid     = errors.New("invalid template format")
	ErrTemplateFieldEmpty  = errors.New("template has no fields")
	ErrExcelEmpty          = errors.New("excel has no data")
)