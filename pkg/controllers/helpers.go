package controllers

import (
	"bytes"
	"io"
	"mime/multipart"
)

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
