package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const MaxUploadSize int64 = 25 << 20 //25mib

var (
	ErrFileRequired        = errors.New("file is required")
	ErrFileTooLarge        = errors.New("file exceeds maximum allowed size")
	ErrUnsupportedFileType = errors.New("unsupported file type")
)

type Service struct {
	repository *Repository
	uploadDir  string
}

func NewService(repository *Repository, uploadDir string) *Service {
	return &Service{
		repository: repository,
		uploadDir:  uploadDir,
	}
}

func (s *Service) Upload(
	ctx context.Context,
	file multipart.File,
	header *multipart.FileHeader,
) (*Document, error) {
	if file == nil || header == nil {
		return nil, ErrFileRequired
	}

	if header.Size <= 0 {
		return nil, ErrFileRequired
	}

	if header.Size > MaxUploadSize {
		return nil, ErrFileTooLarge
	}

	originalName := sanitizeFilename(header.Filename)

	extension := strings.ToLower(filepath.Ext(originalName))

	if !isAllowedExtension(extension) {
		return nil, ErrUnsupportedFileType
	}

	mimeType, err := detectMimeType(file)
	if err != nil {
		return nil, fmt.Errorf("detect file type: %w", err)
	}

	if !isAllowedMimeType(extension, mimeType) {
		return nil, ErrUnsupportedFileType
	}

	documentID := uuid.New()
	storedName := documentID.String() + extension

	if err := os.MkdirAll(s.uploadDir, 0750); err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}

	storagePath := filepath.Join(s.uploadDir, storedName)

	destination, err := os.OpenFile(
		storagePath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0640,
	)
	/*
		O_EXCL -> create this file if not exists
		If somehow this UUID filename already exists:he operation fails rather than overwriting an existing document.

	*/

	if err != nil {
		return nil, fmt.Errorf("create destination file: %w", err)
	}

	defer destination.Close()

	limitedReader := io.LimitReader(file, MaxUploadSize+1)

	written, err := io.Copy(destination, limitedReader)
	if err != nil {
		_ = os.Remove(storagePath)
		return nil, fmt.Errorf("save uploaded file: %w", err)
	}

	if written > MaxUploadSize {
		return nil, ErrFileTooLarge
	}

	doc := &Document{
		ID:           documentID,
		OriginalName: originalName,
		StoredName:   storedName,
		MimeType:     mimeType,
		FileSize:     written,
		StoragePath:  storagePath,
		Status:       StatusUploaded,
	}

	if err := s.repository.Created(ctx, doc); err != nil {
		_ = os.Remove(storagePath)
		return nil, fmt.Errorf("create document metadata: %w", err)
	}

	return doc, nil
}

func sanitizeFilename(filename string) string {
	filename = strings.ReplaceAll(
		filename,
		"\\",
		"/",
	)

	filename = path.Base(filename)
	filename = strings.TrimSpace(filename)

	if filename == "." || filename == "/" || filename == "" {
		return "document"
	}

	return filename
}

func isAllowedExtension(extension string) bool {
	switch extension {
	case ".pdf", ".txt", ".md":
		return true
	default:
		return false
	}
}

func detectMimeType(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)

	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buffer[:n]), nil
}

func isAllowedMimeType(extension string, mimeType string) bool {
	switch extension {
	case ".pdf":
		return mimeType == "application/pdf"
	case ".txt", ".md":
		return strings.HasPrefix(mimeType, "text/plain")
	default:
		return false
	}
}
