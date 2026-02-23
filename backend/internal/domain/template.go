package domain

type Template struct {
	Path   string
	Fields []string
}

func NewTemplate(path string, fields []string) *Template {
	return &Template{
		Path:   path,
		Fields: fields,
	}
}

func (t *Template) Validate() error {
	if len(t.Fields) == 0 {
		return ErrTemplateFieldEmpty
	}
	return nil
}

func (t *Template) ValidateHeaders(headers []string) error {
	headerMap := make(map[string]bool)
	for _, h := range headers {
		headerMap[h] = true
	}

	for _, field := range t.Fields {
		if !headerMap[field] {
			return ErrMissingColumns
		}
	}
	return nil
}