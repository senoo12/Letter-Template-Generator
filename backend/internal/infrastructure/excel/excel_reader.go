package excel

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Reader interface {
	Read(path string) ([]map[string]string, []string, error)
}

type ExcelReader struct {
}

func NewExcelReader() Reader {
	return &ExcelReader{}
}

func (e *ExcelReader) Read(path string) ([]map[string]string, []string, error) {

	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, nil, fmt.Errorf("no sheet found")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}

	if len(rows) < 2 {
		return nil, nil, fmt.Errorf("excel must contain header and at least one row")
	}

	headers := rows[0]
	headers = normalizeHeaders(headers)

	var result []map[string]string

	for _, row := range rows[1:] {

		rowMap := make(map[string]string)

		for i, header := range headers {
			if i < len(row) {
				rowMap[header] = row[i]
			} else {
				rowMap[header] = ""
			}
		}

		result = append(result, rowMap)
	}

	return result, headers, nil
}

func normalizeHeaders(headers []string) []string {
	var normalized []string

	for _, h := range headers {
		h = strings.ToLower(strings.TrimSpace(h))
		h = strings.ReplaceAll(h, " ", "_")
		normalized = append(normalized, h)
	}

	return normalized
}

