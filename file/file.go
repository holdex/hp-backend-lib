package libfile

import (
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrFileTooBig         = errors.New("file too big")
	ErrFileIsEmpty        = errors.New("file is empty")
	ErrWidthTooSmall      = errors.New("width too small")
	ErrHeightTooSmall     = errors.New("height too small")
	ErrWidthNotMatch      = errors.New("width not match")
	ErrHeightNotMatch     = errors.New("height not match")
	ErrFileTypeNotAllowed = errors.New("file type not allowed")
)

func ParseMultipart(w http.ResponseWriter, r *http.Request, fileName string, maxFileSize, maxInMemSize int64) (multipart.File, *multipart.FileHeader, error) {
	r.Body = &maxBytesReader{w: w, r: r.Body, n: maxFileSize}

	if err := r.ParseMultipartForm(maxInMemSize); err != nil {
		return nil, nil, err
	}

	file, header, err := r.FormFile(fileName)
	if err != nil {
		return nil, nil, err
	} else if file == nil {
		return nil, nil, ErrFileIsEmpty
	}

	return file, header, nil
}

type maxBytesReader struct {
	w   http.ResponseWriter
	r   io.ReadCloser // underlying reader
	n   int64         // max bytes remaining
	err error         // sticky error
}

func (l *maxBytesReader) Read(p []byte) (n int, err error) {
	if l.err != nil {
		return 0, l.err
	}
	if len(p) == 0 {
		return 0, nil
	}
	if int64(len(p)) > l.n+1 {
		p = p[:l.n+1]
	}
	n, err = l.r.Read(p)
	if int64(n) <= l.n {
		l.n -= int64(n)
		l.err = err
		return n, err
	}
	n = int(l.n)
	l.n = 0
	type requestTooLarger interface {
		requestTooLarge()
	}
	if res, ok := l.w.(requestTooLarger); ok {
		res.requestTooLarge()
	}
	l.err = ErrFileTooBig
	return n, l.err
}

func (l *maxBytesReader) Close() error {
	return l.r.Close()
}

func IsFileType(file multipart.File, t string) (bool, error) {
	// Create a buffer to store the header of the file in
	fileHeader := make([]byte, 512)

	// Copy the headers into the FileHeader buffer
	if _, err := file.Read(fileHeader); err != nil {
		return false, err
	}

	// set position back to start.
	if _, err := file.Seek(0, 0); err != nil {
		return false, err
	}

	mimeTypes := strings.Split(http.DetectContentType(fileHeader), "/")
	if len(mimeTypes) > 0 && mimeTypes[0] == t {
		return true, nil
	}
	return false, nil
}

func IsMimeType(file multipart.File, types ...string) (bool, error) {
	// Create a buffer to store the header of the file in
	fileHeader := make([]byte, 512)

	// Copy the headers into the FileHeader buffer
	if _, err := file.Read(fileHeader); err != nil {
		return false, err
	}

	// set position back to start.
	if _, err := file.Seek(0, 0); err != nil {
		return false, err
	}

	mimeType := http.DetectContentType(fileHeader)
	if mimeType != "" {
		for _, val := range types {
			if mimeType == val {
				return true, nil

			}
		}
	}

	return false, ErrFileTypeNotAllowed
}

func DecodeImage(file io.Reader, minWidth, minHeight int, formats ...string) (originalImage image.Image, ext string, err error) {
	originalImage, ext, err = image.Decode(file)
	if err != nil {
		return
	}

	if originalImage == nil {
		err = ErrFileIsEmpty
		return
	} else if originalImage.Bounds().Max.X < minWidth {
		err = ErrWidthTooSmall
		return
	} else if originalImage.Bounds().Max.Y < minHeight {
		err = ErrHeightTooSmall
		return
	}

	for _, val := range formats {
		if ext == val {
			return
		}
	}

	err = ErrFileTypeNotAllowed
	return
}

func EncodeImage(originalImage image.Image, ext, fileDir, fileName string) (err error) {
	os.MkdirAll(fileDir, os.ModePerm)
	destinationFile, err := os.OpenFile(filepath.Join(fileDir, fileName), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer destinationFile.Close()

	//Copy original image into the destination
	switch ext {
	case "png":
		{
			err = png.Encode(destinationFile, originalImage)
		}
	case "jpeg":
		{
			err = jpeg.Encode(destinationFile, originalImage, nil)
		}
	}
	return
}

func SaveFileToPath(file io.Reader, fileDir, fileName string) (err error) {
	os.MkdirAll(fileDir, os.ModePerm)
	destinationFile, err := os.OpenFile(filepath.Join(fileDir, fileName), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, file)
	return
}

func DeleteFile(filePath string) error {
	return os.Remove(filePath)
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
