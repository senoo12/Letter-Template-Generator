package request

type GenerateRequest struct {
	TemplateField string
	ExcelField    string
}

func NewGenerateRequest() *GenerateRequest {
	return &GenerateRequest{
		TemplateField: "template",
		ExcelField:    "excel",
	}
}