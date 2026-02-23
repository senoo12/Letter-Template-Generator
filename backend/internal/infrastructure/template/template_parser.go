package template

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Parser interface {
	ExtractFields(path string) ([]string, error)
	Render(path string, data map[string]string) ([]byte, error)
}

type TemplateParser struct{}

func NewTemplateParser() Parser {
	return &TemplateParser{}
}

func (t *TemplateParser) ExtractFields(path string) ([]string, error) {
	content, err := t.readDocumentXML(path)
	if err != nil {
		return nil, err
	}

	// regex untuk menangani tag xml di tengah placeholder
	re := regexp.MustCompile(`\{\{(?:<[^>]*>|\s)*([^{}]+?)(?:<[^>]*>|\s)*\}\}`)
	
	matches := re.FindAllStringSubmatch(content, -1)
	fieldMap := make(map[string]struct{})

	for _, m := range matches {
		if len(m) > 1 {
			// bersihkan dari tag xml dan spasi
			cleanField := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(m[1], "")
			cleanField = strings.TrimSpace(cleanField)
			if cleanField != "" {
				fieldMap[cleanField] = struct{}{}
			}
		}
	}

	if len(fieldMap) == 0 {
		return nil, fmt.Errorf("template has no fields (placeholders like {{name}} not found)")
	}

	var fields []string
	for k := range fieldMap {
		fields = append(fields, k)
	}
	return fields, nil
}

// render melakukan replace placeholder dengan data menggunakan string manipulation
func (t *TemplateParser) Render(templatePath string, data map[string]string) ([]byte, error) {
	// Buka ZIP
	r, err := zip.OpenReader(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template: %w", err)
	}
	defer r.Close()

	// buffer untuk ZIP baru
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// normalisasi data: lowercase keys
	normalizedData := make(map[string]string)
	for k, v := range data {
		normalizedData[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}

		// jika document.xml, lakukan replacement
		if f.Name == "word/document.xml" {
			xmlContent := string(content)
			
			// extract fields yang ada di template
			templateFields, _ := t.ExtractFields(templatePath)
			
			// replace setiap field
			for _, field := range templateFields {
				// buat regex untuk mencari placeholder ini (dengan toleransi tag XML)
				escapedField := regexp.QuoteMeta(field)
				// pattern
				pattern := fmt.Sprintf(`\{\{(?:<[^>]*>|\s)*%s(?:<[^>]*>|\s)*\}\}`, escapedField)
				re := regexp.MustCompile(pattern)
				
				lowerField := strings.ToLower(field)
				if value, ok := normalizedData[lowerField]; ok {
					xmlContent = re.ReplaceAllString(xmlContent, value)
				} else {
					xmlContent = re.ReplaceAllString(xmlContent, "")
				}
			}
			
			content = []byte(xmlContent)
		}

		// menulis ke zip baru
		fw, err := w.Create(f.Name)
		if err != nil {
			return nil, err
		}
		if _, err := fw.Write(content); err != nil {
			return nil, err
		}
	}

	w.Close()
	return buf.Bytes(), nil
}

func (t *TemplateParser) readDocumentXML(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()
			
			b, err := io.ReadAll(rc)
			if err != nil {
				return "", err
			}
			return string(b), nil
		}
	}
	
	return "", fmt.Errorf("document.xml not found")
}