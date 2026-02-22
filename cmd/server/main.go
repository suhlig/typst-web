package main

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"typstweb/typst"
)

//go:generate go run ../typst-fetch

//go:embed index.html
var indexHTML string

func main() {
	err := mainE()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s\n", err)
		os.Exit(1)
	}
}

func mainE() error {
	typstPath := filepath.Join(os.Getenv("KO_DATA_PATH"), "typst", runtime.GOOS, runtime.GOARCH, "typst")
	typst, err := typst.NewTypst(typstPath)

	if err != nil {
		return err
	}

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

		pdf, err := typst.Compile(body)

		if err != nil {
			http.Error(w, fmt.Sprintf("unable to read result PDF: %s", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Write(pdf)
	})

	return http.ListenAndServe(":8080", nil)
}
