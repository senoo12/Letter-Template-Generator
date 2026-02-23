package pdf

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type LibreOfficeGenerator struct{}

func NewLibreOfficeGenerator() Generator {
	return &LibreOfficeGenerator{}
}

func (g *LibreOfficeGenerator) Generate(docxPath, pdfPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// memastikan direktori output ada
	outputDir := filepath.Dir(pdfPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	// absolute paths
	absDocxPath, err := filepath.Abs(docxPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for docx: %w", err)
	}
	
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for output: %w", err)
	}

	// profile unik untuk kibreOffice 
	tempProfile := filepath.Join(os.TempDir(), fmt.Sprintf("lo_profile_%d", time.Now().UnixNano()))
	defer os.RemoveAll(tempProfile)

	// cek file exists
	if _, err := os.Stat(absDocxPath); os.IsNotExist(err) {
		return fmt.Errorf("docx file not found: %s", absDocxPath)
	}

	cmd := exec.CommandContext(ctx,
		"libreoffice",
		"--headless",
		"--convert-to", "pdf:writer_pdf_Export",
		"--outdir", absOutputDir,
		"-env:UserInstallation=file://"+tempProfile, 
		absDocxPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("libreoffice failed: %v, output: %s", err, string(output))
	}

	// verifikasi dan rename PDF
	expectedPdfName := filepath.Base(absDocxPath[:len(absDocxPath)-5]) + ".pdf"
	expectedPdfPath := filepath.Join(absOutputDir, expectedPdfName)
	
	if _, err := os.Stat(expectedPdfPath); os.IsNotExist(err) {
		// cari file PDF yang dihasilkan
		entries, _ := os.ReadDir(absOutputDir)
		for _, entry := range entries {
			if filepath.Ext(entry.Name()) == ".pdf" {
				src := filepath.Join(absOutputDir, entry.Name())
				if src != pdfPath {
					if err := os.Rename(src, pdfPath); err != nil {
						return fmt.Errorf("failed to rename pdf: %w", err)
					}
				}
				return nil
			}
		}
		return fmt.Errorf("pdf file not created, output: %s", string(output))
	}

	// rename ke target jika perlu
	if expectedPdfPath != pdfPath {
		return os.Rename(expectedPdfPath, pdfPath)
	}

	return nil
}