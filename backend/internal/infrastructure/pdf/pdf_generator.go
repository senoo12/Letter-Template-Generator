package pdf

type Generator interface {
    Generate(docxPath, pdfPath string) error
}
