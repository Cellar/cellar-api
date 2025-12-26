package controllers

import (
	"bytes"
	"io"
	"mime/multipart"
)

// FileToBytes reads a multipart file header and returns its contents as a byte slice.
// The file is automatically closed after reading.
func FileToBytes(header *multipart.FileHeader) ([]byte, error) {
	var file multipart.File
	file, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, file); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
