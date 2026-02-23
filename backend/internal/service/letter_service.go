package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"letter-template-generator/internal/domain"
	"letter-template-generator/internal/infrastructure/excel"
	"letter-template-generator/internal/infrastructure/pdf"
	"letter-template-generator/internal/infrastructure/template"
	"letter-template-generator/internal/infrastructure/zipper"
)

type LetterService struct {
	excelReader    excel.Reader
	templateParser template.Parser
	pdfGenerator   pdf.Generator
	zipWriter      zipper.Writer
	fileNameGen    domain.FileNameGenerator
}

func NewLetterService(
	excelReader excel.Reader,
	templateParser template.Parser,
	pdfGenerator pdf.Generator,
	zipWriter zipper.Writer,
) *LetterService {
	return &LetterService{
		excelReader:    excelReader,
		templateParser: templateParser,
		pdfGenerator:   pdfGenerator,
		zipWriter:      zipWriter,
		fileNameGen:    domain.NewFileNameDomain(),
	}
}

func (s *LetterService) Generate(templatePath, excelPath, baseField string) (string, error) {
	// validasi input files exist
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return "", fmt.Errorf("template file not found: %s", templatePath)
	}
	if _, err := os.Stat(excelPath); os.IsNotExist(err) {
		return "", fmt.Errorf("excel file not found: %s", excelPath)
	}

	// nama template untuk generate nama file
	templateFileName := filepath.Base(templatePath)

	// extract template fields
	templateFields, err := s.templateParser.ExtractFields(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to extract template fields: %w", err)
	}

	templateDomain := domain.NewTemplate(templatePath, templateFields)
	if err := templateDomain.Validate(); err != nil {
		return "", err
	}

	// read excel
	rows, headers, err := s.excelReader.Read(excelPath)
	if err != nil {
		return "", fmt.Errorf("failed to read excel: %w", err)
	}

	excelDomain := domain.NewExcelData(excelPath, headers, rows)
	if err := excelDomain.Validate(); err != nil {
		return "", err
	}

	// validate template excel
	if err := templateDomain.ValidateHeaders(excelDomain.Headers); err != nil {
		return "", err
	}

	normalizedBaseField := strings.ToLower(strings.TrimSpace(baseField))
	
	fieldExists := false
	availableFields := make([]string, 0, len(headers))
	
	for _, header := range headers {
		normalizedHeader := strings.ToLower(strings.TrimSpace(header))
		availableFields = append(availableFields, header)
		
		if normalizedHeader == normalizedBaseField {
			fieldExists = true
		}
	}
	
	if !fieldExists {
		return "", fmt.Errorf(
			"base_field '%s' not found in Excel headers. Available fields: %v",
			baseField,
			availableFields,
		)
	}

	// ceate temp directory untuk letters
	tempDir, err := os.MkdirTemp("", "generated-letters-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// generate letters
	for i, row := range excelDomain.Rows {
		normalizedContext := domain.NormalizeMap(row)

		// validasi base field
		if _, ok := normalizedContext[normalizedBaseField]; !ok {
			return "", fmt.Errorf(
				"row %d does not have value for base_field '%s'. Available in row: %v",
				i+1,
				baseField,
				getKeys(normalizedContext),
			)
		}

		// render docx dengan replacement
		renderedDocx, err := s.templateParser.Render(templatePath, normalizedContext)
		if err != nil {
			return "", fmt.Errorf("failed to render template for row %d: %w", i+1, err)
		}

		// generate nama file menggunakan domain
		fileName := s.generateFileName(normalizedContext, normalizedBaseField, i)

		docxFile := filepath.Join(tempDir, fileName+".docx")
		pdfFile := filepath.Join(tempDir, fileName+".pdf")

		// simpan docx
		if err := os.WriteFile(docxFile, renderedDocx, 0644); err != nil {
			return "", fmt.Errorf("failed to write docx: %w", err)
		}

		// convert ke pdf
		if err := s.pdfGenerator.Generate(docxFile, pdfFile); err != nil {
			return "", fmt.Errorf("failed to generate pdf for %s: %w", fileName, err)
		}

		os.Remove(docxFile)
	}

	// zip result ke temp dir
	zipFileName := s.fileNameGen.GenerateZip(templateFileName)
	zipPath := filepath.Join(os.TempDir(), zipFileName)

	if err := s.zipWriter.Zip(tempDir, zipPath); err != nil {
		return "", fmt.Errorf("failed to create zip: %w", err)
	}

	return zipPath, nil 
}

func (s *LetterService) generateFileName(data map[string]string, baseField string, index int) string {
	baseValue, ok := data[baseField]
	if !ok || baseValue == "" {
		// fallback ke field umum jika base_field kosong di row ini
		fallbackFields := []string{"nama", "name", "id", "no", "nomor"}
		for _, field := range fallbackFields {
			if val, exists := data[field]; exists && val != "" {
				baseValue = val
				break
			}
		}
		// ultimate fallback
		if baseValue == "" {
			baseValue = fmt.Sprintf("letter_%d", index+1)
		}
	}

	// sanitize menggunakan domain
	safeBase := s.fileNameGen.Sanitize(baseValue)
	
	// generate nama file dengan random suffix
	return s.fileNameGen.Generate(safeBase)
}

// helper function untuk get keys dari map
func getKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}