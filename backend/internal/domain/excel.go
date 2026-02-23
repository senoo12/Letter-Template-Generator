package domain

type ExcelData struct {
	Path    string
	Headers []string
	Rows    []map[string]string
}

func NewExcelData(path string, headers []string, rows []map[string]string) *ExcelData {
	return &ExcelData{
		Path:    path,
		Headers: headers,
		Rows:    rows,
	}
}

func (e *ExcelData) Validate() error {
	if len(e.Headers) == 0 {
		return ErrExcelEmpty
	}
	if len(e.Rows) == 0 {
		return ErrExcelEmpty
	}
	return nil
}

func (e *ExcelData) HasHeader(field string) bool {
	for _, h := range e.Headers {
		if h == field {
			return true
		}
	}
	return false
}