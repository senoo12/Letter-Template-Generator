package zipper

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
)

type Writer interface {
	Zip(sourceDir, outputZip string) error
	ZipToBuffer(sourceDir string) (*bytes.Buffer, error) // Tambahkan ini
}

type ZipWriter struct{}

func NewZipWriter() Writer {
	return &ZipWriter{}
}

// zip tetap ada untuk kompatibilitas jika Anda butuh simpan ke disk
func (z *ZipWriter) Zip(sourceDir, outputZip string) error {
	zipFile, err := os.Create(outputZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	return z.internalZip(sourceDir, zipFile)
}

// zip to buffer digunakan untuk flow direct download (memory)
func (z *ZipWriter) ZipToBuffer(sourceDir string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	err := z.internalZip(sourceDir, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (z *ZipWriter) internalZip(sourceDir string, w io.Writer) error {
	archive := zip.NewWriter(w)
	defer archive.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			header.Name = info.Name()
		} else {
			header.Name = relPath
		}
		
		header.Method = zip.Deflate
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})
}