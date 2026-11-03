package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadReturnsStoredFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hello"), 0o644)
	_, h := newServer(dir)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/download?name=notes.txt", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != "hello" {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}
}

func TestDownloadMissingFile(t *testing.T) {
	_, h := newServer(t.TempDir())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/download?name=nope.txt", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestListHidesDotfiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), nil, 0o644)
	os.WriteFile(filepath.Join(dir, ".secret"), nil, 0o644)
	_, h := newServer(dir)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/files", nil))
	if rec.Body.String() != "a.txt\n" {
		t.Fatalf("got %q", rec.Body.String())
	}
}

func TestDownloadRejectsPathTraversal(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "uploads")
	os.MkdirAll(root, 0o755)
	os.WriteFile(filepath.Join(base, "secret.txt"), []byte("top secret"), 0o644)
	os.WriteFile(filepath.Join(root, "sub.txt"), []byte("x"), 0o644)
	_, h := newServer(root)

	for _, name := range []string{
		"../secret.txt",
		"foo/../../secret.txt",
		"/etc/passwd",
		filepath.Join(base, "secret.txt"),
		"uploads/sub.txt",
	} {
		req := httptest.NewRequest("GET", "/download", nil)
		q := req.URL.Query()
		q.Set("name", name)
		req.URL.RawQuery = q.Encode()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("name %q: got %d, want 400", name, rec.Code)
		}
		if rec.Body.String() == "top secret" {
			t.Errorf("name %q leaked a file outside the upload directory", name)
		}
	}
}
