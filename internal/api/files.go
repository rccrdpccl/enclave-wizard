package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileUploadHandler struct {
	enclaveDir string
}

func NewFileUploadHandler(enclaveDir string) *FileUploadHandler {
	return &FileUploadHandler{enclaveDir: enclaveDir}
}

func (h *FileUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	const maxSize = 64 << 20 // 64 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}

	dest := r.FormValue("dest")
	if dest == "" {
		http.Error(w, `missing "dest" field`, http.StatusBadRequest)
		return
	}
	if strings.Contains(dest, "..") || filepath.IsAbs(dest) {
		http.Error(w, "invalid dest path", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	destDir := filepath.Join(h.enclaveDir, "config", filepath.Clean(dest))
	if err := os.MkdirAll(destDir, 0750); err != nil {
		http.Error(w, fmt.Sprintf("failed to create directory: %v", err), http.StatusInternalServerError)
		return
	}

	destPath := filepath.Join(destDir, filepath.Base(header.Filename))

	out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create file: %v", err), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, fmt.Sprintf("failed to write file: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": destPath})
}
