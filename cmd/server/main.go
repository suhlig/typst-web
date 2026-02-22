package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:generate go run ../typst-fetch

//go:embed index.html
var indexHTML string

func main() {
	typstPath := filepath.Join(os.Getenv("KO_DATA_PATH"), "typst", "0.14.2", runtime.GOOS, runtime.GOARCH, "typst")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Write([]byte(indexHTML))
			return
		}

		err := r.ParseForm()

		if err != nil {
			http.Error(w, fmt.Sprintf("could not parse form: %s", err.Error()), 500)
			return
		}

		body := r.FormValue("source")

		pdfPath, err := os.CreateTemp("", "out.pdf")

		cmd := exec.Command(typstPath, "compile", "-", "--format", "pdf", pdfPath.Name())

		stdin, err := cmd.StdinPipe()

		if err != nil {
			http.Error(w, fmt.Sprintf("unable to get stdin pipe: %s", err.Error()), 500)
			return
		}

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		go func() {
			stdin.Write([]byte(body))
			stdin.Close()
		}()

		err = cmd.Run()

		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to run command %v: %s\n", cmd, err)
			fmt.Fprintf(os.Stderr, "STDOUT:\n%s\n", stdout.String())
			fmt.Fprintf(os.Stderr, "STDERR:\n%s\n", stderr.String())
			http.Error(w, "failed to run command", 500)
			return
		}

		pdf, err := os.ReadFile(pdfPath.Name())

		if err != nil {
			http.Error(w, fmt.Sprintf("unable to read result PDF: %s", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Write(pdf)
	})

	http.ListenAndServe(":8080", nil)
}
