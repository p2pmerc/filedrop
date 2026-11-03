// filedrop is a tiny file sharing server: upload a file, share its name, download it.
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type server struct {
	root string
}

func (s *server) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	name := filepath.Base(header.Filename)
	out, err := os.Create(filepath.Join(s.root, name))
	if err != nil {
		http.Error(w, "cannot store file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "cannot store file", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "stored %s\n", name)
}

func (s *server) download(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	path := filepath.Join(s.root, name)
	data, err := os.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(name)))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

func (s *server) list(w http.ResponseWriter, _ *http.Request) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		http.Error(w, "cannot list files", http.StatusInternalServerError)
		return
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	fmt.Fprintln(w, strings.Join(names, "\n"))
}

func newServer(root string) (*server, http.Handler) {
	s := &server{root: root}
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", s.upload)
	mux.HandleFunc("/download", s.download)
	mux.HandleFunc("/files", s.list)
	return s, mux
}

func main() {
	root := os.Getenv("FILEDROP_DIR")
	if root == "" {
		root = "./uploads"
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		log.Fatal(err)
	}
	_, handler := newServer(root)
	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	log.Printf("filedrop serving %s on %s", root, addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
