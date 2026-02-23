package domain

type LetterValidator struct{}

func NewLetterValidator() *LetterValidator {
	return &LetterValidator{}
}

func (v *LetterValidator) ValidateTemplateAndExcel(t *Template, e *ExcelData) error {
	if err := t.Validate(); err != nil {
		return err
	}

	if err := e.Validate(); err != nil {
		return err
	}

	return t.ValidateHeaders(e.Headers)
}