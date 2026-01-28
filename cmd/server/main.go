package main

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed index.html
var indexHTML string

//go:embed ../../internal/embedded/typst/typst
var typstBin []byte

func extractTypst() (string, error) {
	cache, _ := os.UserCacheDir()
	dir := filepath.Join(cache, "typstweb")
	os.MkdirAll(dir, 0755)

	path := filepath.Join(dir, "typst")
	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, typstBin, 0755); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		return "", err
	}
	return path, nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Write([]byte(indexHTML))
			return
		}

		body, _ := io.ReadAll(r.Body)
		typstPath, err := extractTypst()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		cmd := exec.Command(typstPath, "-", "out.pdf")
		stdin, _ := cmd.StdinPipe()

		go func() {
			stdin.Write(body)
			stdin.Close()
		}()

		if err := cmd.Run(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		pdf, _ := os.ReadFile("out.pdf")
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(pdf)
	})

	http.ListenAndServe(":8080", nil)
}
