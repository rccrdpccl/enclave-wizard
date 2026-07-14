package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

func setupFilesAPI(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	enclaveDir := t.TempDir()
	configDir := filepath.Join(enclaveDir, "config")
	os.MkdirAll(configDir, 0755)

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("test", "0.0.0"))
	NewFileUploadHandler(enclaveDir).Register(api)
	return httptest.NewServer(mux), enclaveDir
}

func createMultipartRequest(t *testing.T, url string, fields map[string]string, fileName, fileContent string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("writing field %s: %v", k, err)
		}
	}

	if fileName != "" {
		fw, err := w.CreateFormFile("file", fileName)
		if err != nil {
			t.Fatalf("creating form file: %v", err)
		}
		if _, err := io.WriteString(fw, fileContent); err != nil {
			t.Fatalf("writing file content: %v", err)
		}
	}

	w.Close()

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestFileUpload_Success(t *testing.T) {
	srv, enclaveDir := setupFilesAPI(t)
	defer srv.Close()

	req := createMultipartRequest(t, srv.URL+"/api/v1/files",
		map[string]string{"dest": "certs"},
		"server.crt", "-----BEGIN CERTIFICATE-----\nfake\n-----END CERTIFICATE-----\n",
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d, body: %s", resp.StatusCode, http.StatusOK, string(body))
	}

	var result struct {
		Path string `json:"path"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	expectedPath := filepath.Join(enclaveDir, "config", "certs", "server.crt")
	if result.Path != expectedPath {
		t.Errorf("path = %q, want %q", result.Path, expectedPath)
	}

	// Verify file exists and has correct content
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("reading uploaded file: %v", err)
	}
	if !strings.Contains(string(content), "BEGIN CERTIFICATE") {
		t.Errorf("file content = %q, expected certificate content", string(content))
	}
}

func TestFileUpload_MissingDest(t *testing.T) {
	srv, _ := setupFilesAPI(t)
	defer srv.Close()

	req := createMultipartRequest(t, srv.URL+"/api/v1/files",
		map[string]string{},
		"test.txt", "content",
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestFileUpload_MissingFile(t *testing.T) {
	srv, _ := setupFilesAPI(t)
	defer srv.Close()

	// Send multipart with only dest field, no file
	req := createMultipartRequest(t, srv.URL+"/api/v1/files",
		map[string]string{"dest": "certs"},
		"", "", // no file
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestFileUpload_PathTraversal(t *testing.T) {
	srv, _ := setupFilesAPI(t)
	defer srv.Close()

	cases := []struct {
		name string
		dest string
	}{
		{"dotdot", "../etc"},
		{"embedded dotdot", "certs/../../etc"},
		{"absolute path", "/etc/passwd"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := createMultipartRequest(t, srv.URL+"/api/v1/files",
				map[string]string{"dest": tc.dest},
				"evil.txt", "malicious",
			)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("dest=%q: status = %d, want %d", tc.dest, resp.StatusCode, http.StatusBadRequest)
			}
		})
	}
}

func TestFileUpload_NestedDest(t *testing.T) {
	srv, enclaveDir := setupFilesAPI(t)
	defer srv.Close()

	req := createMultipartRequest(t, srv.URL+"/api/v1/files",
		map[string]string{"dest": "deep/nested/dir"},
		"data.json", `{"key":"value"}`,
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want %d, body: %s", resp.StatusCode, http.StatusOK, string(body))
	}

	expectedPath := filepath.Join(enclaveDir, "config", "deep", "nested", "dir", "data.json")
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("reading uploaded file: %v", err)
	}
	if string(content) != `{"key":"value"}` {
		t.Errorf("file content = %q, want %q", string(content), `{"key":"value"}`)
	}
}

func TestFileUpload_AppearsInOpenAPISpec(t *testing.T) {
	mux := http.NewServeMux()
	cfg := huma.DefaultConfig("test", "0.0.0")
	api := humago.New(mux, cfg)
	NewFileUploadHandler(t.TempDir()).Register(api)

	// Check the OpenAPI spec has the /api/v1/files path
	spec := api.OpenAPI()
	paths := spec.Paths
	if paths == nil {
		t.Fatal("OpenAPI paths is nil")
	}

	filePath := paths["/api/v1/files"]
	if filePath == nil {
		t.Fatal("expected /api/v1/files in OpenAPI paths")
	}

	post := filePath.Post
	if post == nil {
		t.Fatal("expected POST operation on /api/v1/files")
	}

	if post.OperationID != "upload-file" {
		t.Errorf("operationID = %q, want %q", post.OperationID, "upload-file")
	}

	if post.Summary != "Upload a file" {
		t.Errorf("summary = %q, want %q", post.Summary, "Upload a file")
	}

	// Verify it has request body with multipart/form-data content type
	if post.RequestBody == nil {
		t.Fatal("expected request body")
	}
	if _, ok := post.RequestBody.Content["multipart/form-data"]; !ok {
		t.Errorf("expected multipart/form-data content type, got keys: %v", keysOf(post.RequestBody.Content))
	}
}

func keysOf[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
