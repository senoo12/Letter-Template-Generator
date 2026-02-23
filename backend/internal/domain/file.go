package domain

import (
	"math/rand"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

func generateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// sanitizeFileName bersihkan karakter aneh
func SanitizeFileName(name string) string {
	name = strings.ToLower(name)
	name = strings.TrimSpace(name)

	// hapus karakter aneh
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	name = reg.ReplaceAllString(name, "_")

	// hapus underscore beruntun
	name = regexp.MustCompile(`_+`).ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")

	if name == "" {
		return "unnamed"
	}

	return name
}

// generate nama file dengan random suffix
func GenerateDynamicFileName(base string) string {
	return base + "_" + generateRandomString(6)
}

// interface untuk generate nama file
type FileNameGenerator interface {
	Generate(base string) string
	GenerateZip(templateName string) string
	Sanitize(name string) string
}

// implementasi domain nama file
type FileNameDomain struct{}

func NewFileNameDomain() *FileNameDomain {
	return &FileNameDomain{}
}

func (f *FileNameDomain) Generate(base string) string {
	return GenerateDynamicFileName(base)
}

// generate nama file zip dari nama template + random + generated
func (f *FileNameDomain) GenerateZip(templateName string) string {
	// hapus extension .docx atau .doc
	base := strings.TrimSuffix(templateName, filepath.Ext(templateName))
	
	// sanitize nama template
	safeName := SanitizeFileName(base)
	
	// format: nama_template_randomstring_generated.zip
	return safeName + "_" + generateRandomString(6) + "_generated.zip"
}

func (f *FileNameDomain) Sanitize(name string) string {
	return SanitizeFileName(name)
}