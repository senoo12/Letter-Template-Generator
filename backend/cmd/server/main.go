package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"letter-template-generator/internal/handler"
	"letter-template-generator/internal/infrastructure/excel"
	"letter-template-generator/internal/infrastructure/pdf"
	"letter-template-generator/internal/infrastructure/template"
	"letter-template-generator/internal/infrastructure/zipper"
	"letter-template-generator/internal/service"
)

func main() {

	// set up gin
	router := gin.Default()
	router.MaxMultipartMemory = 20 << 20

	// infrstruktur
	excelReader := excel.NewExcelReader()
	templateParser := template.NewTemplateParser()
	pdfGenerator := pdf.NewLibreOfficeGenerator()
	zipWriter := zipper.NewZipWriter()

	// service
	letterService := service.NewLetterService(
		excelReader,
		templateParser,
		pdfGenerator,
		zipWriter,
	)

	// handler
	letterHandler := handler.NewLetterHandler(letterService)

	// routes
	router.POST("/generate", letterHandler.Generate)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/downloads/:filename", letterHandler.Download)

	// server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Printf("Temp directory: %s", os.TempDir())
	
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}