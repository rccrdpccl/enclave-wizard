package api

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// FileUploadHandler handles file uploads to the enclave config directory.
type FileUploadHandler struct {
	enclaveDir string
}

// NewFileUploadHandler creates a new FileUploadHandler.
func NewFileUploadHandler(enclaveDir string) *FileUploadHandler {
	return &FileUploadHandler{enclaveDir: enclaveDir}
}

// FileUploadInput is the huma input type for multipart file uploads.
// The RawBody field tells huma this is a multipart/form-data request.
type FileUploadInput struct {
	RawBody multipart.Form
}

// FileUploadOutput is the huma output type for file uploads.
type FileUploadOutput struct {
	Body struct {
		Path string `json:"path" doc:"Absolute path where the file was written"`
	}
}

const maxUploadSize = 64 << 20 // 64 MB

// uploadFile handles a file upload request.
func (h *FileUploadHandler) uploadFile(_ context.Context, input *FileUploadInput) (*FileUploadOutput, error) {
	// Extract dest field
	destValues := input.RawBody.Value["dest"]
	if len(destValues) == 0 || destValues[0] == "" {
		return nil, huma.Error400BadRequest(`missing "dest" field`)
	}
	dest := destValues[0]

	if strings.Contains(dest, "..") || filepath.IsAbs(dest) {
		return nil, huma.Error400BadRequest("invalid dest path")
	}

	// Extract uploaded file
	fileHeaders := input.RawBody.File["file"]
	if len(fileHeaders) == 0 {
		return nil, huma.Error400BadRequest("missing file")
	}
	header := fileHeaders[0]

	// Check size
	if header.Size > maxUploadSize {
		return nil, huma.Error400BadRequest(fmt.Sprintf("file too large (max %d MB)", maxUploadSize>>20))
	}

	file, err := header.Open()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to open uploaded file", err)
	}
	defer file.Close()

	destDir := filepath.Join(h.enclaveDir, "config", filepath.Clean(dest))
	if err := os.MkdirAll(destDir, 0750); err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("failed to create directory: %v", err))
	}

	destPath := filepath.Join(destDir, filepath.Base(header.Filename))

	out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("failed to create file: %v", err))
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return nil, huma.Error500InternalServerError(fmt.Sprintf("failed to write file: %v", err))
	}

	result := &FileUploadOutput{}
	result.Body.Path = destPath
	return result, nil
}

// Register registers the file upload endpoint with the huma API.
func (h *FileUploadHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "upload-file",
		Method:        http.MethodPost,
		Path:          "/api/v1/files",
		Summary:       "Upload a file",
		Description:   "Upload a file to the enclave config directory. The file is written to config/<dest>/<filename>.",
		Tags:          []string{"Files"},
		MaxBodyBytes:  maxUploadSize,
		DefaultStatus: http.StatusOK,
	}, h.uploadFile)
}
