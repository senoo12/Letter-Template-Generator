package response

type GenerateResponse struct {
	FileName    string `json:"file_name"`
	DownloadURL string `json:"download_url"`
}