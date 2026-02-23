package handler

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"letter-template-generator/internal/domain"
	requestDTO "letter-template-generator/internal/dto/request"
	responseDTO "letter-template-generator/internal/dto/response"
	"letter-template-generator/internal/service"
	"letter-template-generator/pkg/response"
)

type LetterHandler struct {
	service *service.LetterService
}

func NewLetterHandler(s *service.LetterService) *LetterHandler {
	return &LetterHandler{
		service: s,
	}
}

func (h *LetterHandler) Generate(c *gin.Context) {
	req := requestDTO.NewGenerateRequest()

	// ambil base field
	baseField := c.PostForm("base_file_field")
	if baseField == "" {
		response.BadRequest(c, "base_file_field is required")
		return
	}

	// validate upload
	templateFile, err := c.FormFile(req.TemplateField)
	if err != nil {
		response.BadRequest(c, "template file is required")
		return
	}

	excelFile, err := c.FormFile(req.ExcelField)
	if err != nil {
		response.BadRequest(c, "excel file is required")
		return
	}

	// create temp upload dir
	tempDir, err := os.MkdirTemp("", "temp-upload-*")
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	defer os.RemoveAll(tempDir)

	// sanitize nama file menggunakan domain
	fileNameDomain := domain.NewFileNameDomain()
	cleanTemplateName := fileNameDomain.Sanitize(templateFile.Filename)
	cleanExcelName := fileNameDomain.Sanitize(excelFile.Filename)

	templatePath := filepath.Join(tempDir, cleanTemplateName)
	excelPath := filepath.Join(tempDir, cleanExcelName)

	// simpan file
	if err := c.SaveUploadedFile(templateFile, templatePath); err != nil {
		response.InternalError(c, "failed to save template")
		return
	}
	if err := c.SaveUploadedFile(excelFile, excelPath); err != nil {
		response.InternalError(c, "failed to save excel")
		return
	}

	// call service - return full path
	zipPath, err := h.service.Generate(templatePath, excelPath, baseField)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// extract filename dari full path
	zipFileName := filepath.Base(zipPath)

	log.Printf("Generated zip: %s at %s", zipFileName, zipPath)

	// response
	res := responseDTO.GenerateResponse{
		FileName:    zipFileName,
		DownloadURL: "/downloads/" + zipFileName,
	}

	response.OK(c, res)
}

// download handler untuk serve file dari temp dir
func (h *LetterHandler) Download(c *gin.Context) {
	filename := c.Param("filename")
	
	log.Printf("Download request for: %s", filename)
	
	// cek apakah sudah berakhiran .zip (case insensitive)
	lowerFilename := strings.ToLower(filename)
	if !strings.HasSuffix(lowerFilename, ".zip") {
		filename += ".zip"
	}
	
	// mencari di temp dir - langsung pakai filename tanpa sanitize ulang
	filePath := filepath.Join(os.TempDir(), filename)
	
	log.Printf("Looking for file at: %s", filePath)
	
	// verifikasi file exists dan bukan directory
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fileNameDomain := domain.NewFileNameDomain()
			sanitizedFilename := fileNameDomain.Sanitize(filename)
			if !strings.HasSuffix(sanitizedFilename, ".zip") {
				sanitizedFilename += ".zip"
			}
			filePath = filepath.Join(os.TempDir(), sanitizedFilename)
			
			log.Printf("Retry with sanitized: %s", filePath)
			
			info, err = os.Stat(filePath)
			if err != nil {
				log.Printf("File not found: %s", filePath)
				c.JSON(404, gin.H{
					"error": "file not found", 
					"requested": filename,
					"paths_checked": []string{
						filepath.Join(os.TempDir(), filename),
						filepath.Join(os.TempDir(), sanitizedFilename),
					},
				})
				return
			}
		} else {
			log.Printf("Stat error: %v", err)
			c.JSON(500, gin.H{"error": "internal error"})
			return
		}
	}
	
	if info.IsDir() {
		c.JSON(400, gin.H{"error": "invalid file"})
		return
	}
	
	log.Printf("Serving file: %s (%d bytes)", filePath, info.Size())
	
	// serve file dengan proper headers untuk download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))
	c.Header("Content-Type", "application/zip")
	c.File(filePath)
}