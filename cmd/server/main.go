package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

//go:generate go run ../typst-fetch

//go:embed index.html
var indexHTML string

func extractTypst() (string, error) {
	cache, err := os.UserCacheDir()

	if err != nil {
		return "", err
	}

	dir := filepath.Join(cache, "typstweb")
	err = os.MkdirAll(dir, 0755)

	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, "typst")
	tmp := path + ".tmp"

	err = os.WriteFile(tmp, typstBin, 0755)

	if err != nil {
		return "", err
	}

	err = os.Rename(tmp, path)

	if err != nil {
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

		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, fmt.Sprintf("could not read body: %s", err.Error()), 500)
			return
		}

		typstPath, err := extractTypst()

		if err != nil {
			http.Error(w, fmt.Sprintf("unable to extract typst: %s", err.Error()), 500)
			return
		}

		cmd := exec.Command(typstPath, "compile", "-", "out.pdf")

		stdin, err := cmd.StdinPipe()

		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get stdin pipe: %s", err.Error()), 500)
			return
		}

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		go func() {
			stdin.Write(body)
			stdin.Close()
		}()

		err = cmd.Run()

		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to run command: %s", stderr.String())
			http.Error(w, "failed to run command", 500)
			return
		}

		pdf, err := os.ReadFile("out.pdf")

		if err != nil {
			http.Error(w, fmt.Sprintf("unable to read result PDF: %s", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Write(pdf)
	})

	http.ListenAndServe(":8080", nil)
}
