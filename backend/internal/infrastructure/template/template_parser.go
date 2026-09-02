package template

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	"golang.org/x/image/webp"
	"io"
	"log"
	"net/http"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Parser interface {
	ExtractFields(path string) ([]string, error)
	Render(path string, data map[string]string) ([]byte, error)
}

type TemplateParser struct{}

func NewTemplateParser() Parser {
	return &TemplateParser{}
}

// tagOrSpace adalah potongan regex yang mentolerir tag xml (dari run split
// Word) atau spasi di antara karakter/token.
const tagOrSpace = `(?:<[^>]*>|\s)*`

func (t *TemplateParser) ExtractFields(path string) ([]string, error) {
	content, err := t.readDocumentXML(path)
	if err != nil {
		return nil, err
	}

	// PENTING: dua kurung kurawal pembuka/penutup ("{{" dan "}}") sendiri
	// bisa terpecah jadi run terpisah oleh Word (mis. "{<run break>{ttd}}").
	// Maka setiap karakter kurung diberi toleransi tagOrSpace di antaranya,
	// bukan cuma di sekitar keseluruhan "{{...}}".
	re := regexp.MustCompile(
		`\{` + tagOrSpace + `\{` + tagOrSpace + `([^{}]+?)` + tagOrSpace + `\}` + tagOrSpace + `\}`,
	)

	matches := re.FindAllStringSubmatch(content, -1)
	fieldMap := make(map[string]struct{})

	for _, m := range matches {
		if len(m) > 1 {
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

// buildPlaceholderPattern membangun regex untuk mencari SATU placeholder
// tertentu (mis. field "ttd") di document.xml, dengan toleransi run-split
// Word di SETIAP karakter — termasuk di antara kedua kurung kurawal itu
// sendiri, dan di antara setiap huruf nama field. Ini perlu karena Word bisa
// memecah "{{ttd}}" jadi run-run kecil seperti "{", "{tt", "d}", "}" dsb.
func buildPlaceholderPattern(field string) string {
	var sb strings.Builder
	sb.WriteString(`\{` + tagOrSpace + `\{` + tagOrSpace)
	for _, r := range field {
		sb.WriteString(regexp.QuoteMeta(string(r)))
		sb.WriteString(tagOrSpace)
	}
	sb.WriteString(`\}` + tagOrSpace + `\}`)
	return sb.String()
}

// ==================================================================
// Dukungan gambar
// ==================================================================
//
// Value dari Excel bisa berupa link gambar (http/https diakhiri
// .png/.jpg/.jpeg/.gif/.bmp/.webp). Kalau terdeteksi sebagai link
// gambar, gambar di-download lalu di-embed sebagai <w:drawing> asli
// (bukan teks) - butuh 3 bagian docx yang disentuh sekaligus:
//   1. word/media/imageN.ext        -> file biner gambar (entry baru)
//   2. word/_rels/document.xml.rels -> relationship rId -> media
//   3. [Content_Types].xml          -> registrasi ekstensi file
//   4. word/document.xml            -> placeholder diganti <w:drawing>

const emuPerInch = 914400
const maxImageWidthInch = 3.0 // lebar maksimum gambar di surat; sesuaikan bila perlu

// isURL hanya cek apakah value berbentuk link http(s). TIDAK menebak dari
// ekstensi di URL (rapuh -- banyak CDN seperti Wikia/Fandom, Google Drive,
// Cloudinary taruh ekstensi di tengah path atau tidak menampilkannya sama
// sekali, mis. ".../Naruto_newshot.jpg/revision/latest?cb=...").
// Validitas gambar yang sebenarnya ditentukan lewat isi responsnya di
// fetchImage (Content-Type + berhasil/tidaknya di-decode).
func isURL(v string) bool {
	v = strings.TrimSpace(v)
	return strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://")
}

type downloadedImage struct {
	data      []byte
	ext       string
	widthEMU  int64
	heightEMU int64
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

func fetchImage(url string) (*downloadedImage, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("url tidak valid %s: %w", url, err)
	}
	// beberapa CDN (mis. Wikia/Fandom) menolak request tanpa User-Agent dan/atau
	// menerapkan proteksi hotlink yang butuh Referer dari domain aslinya.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if parsed, perr := neturl.Parse(url); perr == nil && parsed.Host != "" {
		req.Header.Set("Referer", parsed.Scheme+"://"+parsed.Host+"/")
	}
	req.Header.Set("Accept", "image/*,*/*;q=0.8")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download gagal, status %d untuk %s", resp.StatusCode, url)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// validasi SEBENARNYA bahwa ini gambar: coba decode isinya, bukan
	// menebak dari nama/ekstensi URL. Kalau gagal di-decode berarti ini
	// bukan gambar (mis. halaman HTML biasa) -> caller akan fallback ke teks.
	//
	// PNG/JPEG/GIF dicek pakai stdlib (image.DecodeConfig, tanpa decode penuh
	// biar hemat). WebP TIDAK didukung stdlib Go sama sekali, jadi dicoba
	// terpisah pakai golang.org/x/image/webp -- dan karena dukungan WebP di
	// Word/LibreOffice suka tidak konsisten, hasil WebP selalu dikonversi ke
	// PNG dulu (bukan disimpan mentah) supaya aman ditanam ke docx.
	cfg, format, decErr := image.DecodeConfig(bytes.NewReader(data))
	if decErr != nil {
		ct := resp.Header.Get("Content-Type")
		return nil, fmt.Errorf("konten di %s bukan gambar yang bisa dibaca (content-type: %s): %w", url, ct, decErr)
	}

	// dukungan WebP di Word/LibreOffice suka tidak konsisten, jadi kalau
	// formatnya webp, SELALU decode penuh dan konversi ke PNG dulu -- jangan
	// tanam bytes webp mentahnya ke docx.
	if format == "webp" {
		webpImg, werr := webp.Decode(bytes.NewReader(data))
		if werr != nil {
			return nil, fmt.Errorf("gagal decode penuh webp dari %s: %w", url, werr)
		}

		var pngBuf bytes.Buffer
		if encErr := png.Encode(&pngBuf, webpImg); encErr != nil {
			return nil, fmt.Errorf("gagal konversi webp ke png untuk %s: %w", url, encErr)
		}

		bounds := webpImg.Bounds()
		widthInch := maxImageWidthInch
		heightInch := widthInch * float64(bounds.Dy()) / float64(bounds.Dx())

		return &downloadedImage{
			data:      pngBuf.Bytes(),
			ext:       "png",
			widthEMU:  int64(widthInch * emuPerInch),
			heightEMU: int64(heightInch * emuPerInch),
		}, nil
	}

	widthInch := maxImageWidthInch
	heightInch := widthInch * float64(cfg.Height) / float64(cfg.Width)

	return &downloadedImage{
		data:      data,
		ext:       format,
		widthEMU:  int64(widthInch * emuPerInch),
		heightEMU: int64(heightInch * emuPerInch),
	}, nil
}

func normalizeExt(ext string) string {
	if ext == "jpeg" {
		return "jpg"
	}
	return ext
}

func contentTypeForExt(ext string) (string, bool) {
	m := map[string]string{
		"png": "image/png",
		"jpg": "image/jpeg",
		"gif": "image/gif",
		"bmp": "image/bmp",
	}
	v, ok := m[ext]
	return v, ok
}

func nextRelID(relsXML string) int {
	re := regexp.MustCompile(`Id="rId(\d+)"`)
	matches := re.FindAllStringSubmatch(relsXML, -1)
	max := 0
	for _, m := range matches {
		n, _ := strconv.Atoi(m[1])
		if n > max {
			max = n
		}
	}
	return max + 1
}

func addImageRelationship(relsXML string, rID int, target string) string {
	newRel := fmt.Sprintf(
		`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="%s"/>`,
		rID, target,
	)
	return strings.Replace(relsXML, "</Relationships>", newRel+"</Relationships>", 1)
}

func ensureContentType(ctXML string, ext string) string {
	mime, ok := contentTypeForExt(ext)
	if !ok {
		return ctXML
	}
	if strings.Contains(ctXML, fmt.Sprintf(`Extension="%s"`, ext)) {
		return ctXML
	}
	newDefault := fmt.Sprintf(`<Default Extension="%s" ContentType="%s"/>`, ext, mime)
	return strings.Replace(ctXML, "</Types>", newDefault+"</Types>", 1)
}

func buildDrawingXML(rID int, docPrID int, cx, cy int64) string {
	return fmt.Sprintf(
		`<w:r><w:rPr/><w:drawing>`+
			`<wp:inline distT="0" distB="0" distL="0" distR="0" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing">`+
			`<wp:extent cx="%d" cy="%d"/>`+
			`<wp:effectExtent l="0" t="0" r="0" b="0"/>`+
			`<wp:docPr id="%d" name="Picture %d"/>`+
			`<wp:cNvGraphicFramePr><a:graphicFrameLocks xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" noChangeAspect="1"/></wp:cNvGraphicFramePr>`+
			`<a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">`+
			`<a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture">`+
			`<pic:pic xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">`+
			`<pic:nvPicPr><pic:cNvPr id="%d" name="Picture %d"/><pic:cNvPicPr/></pic:nvPicPr>`+
			`<pic:blipFill><a:blip r:embed="rId%d" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill>`+
			`<pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr>`+
			`</pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r>`,
		cx, cy, docPrID, docPrID, docPrID, docPrID, rID, cx, cy,
	)
}

// findEnclosingRunSpan mencari batas <w:r ...>...</w:r> yang membungkus
// sebuah match placeholder. Ini WAJIB dipakai untuk penggantian gambar --
// <w:drawing> adalah elemen struktural yang harus jadi ANAK LANGSUNG dari
// <w:r>, bukan diselipkan mentah di dalam <w:t> (teks). Kalau cuma
// menggantikan teks placeholder di dalam <w:t> seperti substitusi teks
// biasa, hasilnya <w:t><w:r><w:drawing>...</w:drawing></w:r></w:t> --
// XML yang tidak valid secara skema OOXML, sehingga Word/LibreOffice diam-
// diam membuang elemen itu (hasil akhirnya kosong, tanpa error apa pun).
func findEnclosingRunSpan(content string, matchStart, matchEnd int) (int, int) {
	openRe := regexp.MustCompile(`<w:r(?:\s[^>]*)?>`)
	closeRe := regexp.MustCompile(`</w:r>`)

	opens := openRe.FindAllStringIndex(content, -1)
	closes := closeRe.FindAllStringIndex(content, -1)

	runStart := matchStart
	for _, o := range opens {
		if o[0] <= matchStart {
			runStart = o[0]
		} else {
			break
		}
	}

	runEnd := matchEnd
	for _, c := range closes {
		if c[1] >= matchEnd {
			runEnd = c[1]
			break
		}
	}

	return runStart, runEnd
}

// replaceImagePlaceholder mengganti SELURUH <w:r>...</w:r> yang membungkus
// placeholder (bisa lebih dari satu <w:r> kalau placeholder-nya kepecah run
// oleh Word) dengan drawingXML, bukan cuma teksnya saja.
func replaceImagePlaceholder(content string, re *regexp.Regexp, drawingXML string) string {
	for {
		loc := re.FindStringIndex(content)
		if loc == nil {
			break
		}
		runStart, runEnd := findEnclosingRunSpan(content, loc[0], loc[1])
		content = content[:runStart] + drawingXML + content[runEnd:]
	}
	return content
}

func escapeXMLText(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(s)
}

// ==================================================================
// Render
// ==================================================================

func (t *TemplateParser) Render(templatePath string, data map[string]string) ([]byte, error) {
	r, err := zip.OpenReader(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template: %w", err)
	}
	defer r.Close()

	files := make(map[string][]byte)
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
		files[f.Name] = content
	}

	docBytes, ok := files["word/document.xml"]
	if !ok {
		return nil, fmt.Errorf("word/document.xml not found in template")
	}
	xmlContent := string(docBytes)

	const relsPath = "word/_rels/document.xml.rels"
	relsBytes, ok := files[relsPath]
	if !ok {
		return nil, fmt.Errorf("%s not found in template", relsPath)
	}
	relsContent := string(relsBytes)

	const ctPath = "[Content_Types].xml"
	ctBytes, ok := files[ctPath]
	if !ok {
		return nil, fmt.Errorf("%s not found in template", ctPath)
	}
	ctContent := string(ctBytes)

	normalizedData := make(map[string]string)
	for k, v := range data {
		normalizedData[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}

	templateFields, err := t.ExtractFields(templatePath)
	if err != nil {
		return nil, err
	}

	nextRID := nextRelID(relsContent)
	nextDocPrID := 1000

	for _, field := range templateFields {
		// gunakan pattern toleran-run-split, bukan cuma QuoteMeta biasa,
		// karena kurung kurawal & huruf field bisa terpecah run oleh Word
		pattern := buildPlaceholderPattern(field)
		re := regexp.MustCompile(pattern)

		lowerField := strings.ToLower(field)
		value, exists := normalizedData[lowerField]

		if !exists || value == "" {
			xmlContent = re.ReplaceAllString(xmlContent, "")
			continue
		}

		if isURL(value) {
			img, err := fetchImage(value)
			if err != nil {
				// LOG alasan gagalnya supaya kelihatan di log server -- selama ini
				// error ditelan diam-diam sehingga sulit didiagnosis dari luar.
				log.Printf("[template] gagal embed gambar untuk field '%s' (value=%s): %v", field, value, err)
				xmlContent = re.ReplaceAllString(xmlContent, escapeXMLText(value))
				continue
			}
			log.Printf("[template] berhasil embed gambar untuk field '%s' (%dx%d EMU)", field, img.widthEMU, img.heightEMU)

			ext := normalizeExt(img.ext)
			mediaName := fmt.Sprintf("image%d.%s", nextRID, ext)
			mediaPath := "word/media/" + mediaName
			files[mediaPath] = img.data

			relsContent = addImageRelationship(relsContent, nextRID, "media/"+mediaName)
			ctContent = ensureContentType(ctContent, ext)

			drawingXML := buildDrawingXML(nextRID, nextDocPrID, img.widthEMU, img.heightEMU)
			xmlContent = replaceImagePlaceholder(xmlContent, re, drawingXML)

			nextRID++
			nextDocPrID++
			continue
		}

		xmlContent = re.ReplaceAllString(xmlContent, escapeXMLText(value))
	}

	files["word/document.xml"] = []byte(xmlContent)
	files[relsPath] = []byte(relsContent)
	files[ctPath] = []byte(ctContent)

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		fw, err := w.Create(name)
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