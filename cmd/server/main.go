package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/suhlig/gosmos/typst"
)

//go:generate go tool github.com/suhlig/gosmos/cmd/fetch-typst --path kodata/typst
//go:generate go tool github.com/suhlig/gosmos/cmd/fetch-fonts --path kodata/fonts "Meie Script" "Roboto"

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
	koDataPath := os.Getenv("KO_DATA_PATH")

	if koDataPath == "" {
		koDataPath = "kodata" // Fall back to kodata next to the binary, or relative to source
	}

	typstBinary := filepath.Join(koDataPath, "typst", runtime.GOOS, runtime.GOARCH, "typst")

	compiler, err := typst.NewCompileCommand(
		typstBinary,
		typst.WithoutSystemFonts(),
		typst.WithoutEmbeddedFonts(),
		typst.WithFontPath(filepath.Join(koDataPath, "fonts")),
	)

	if err != nil {
		return err
	}

	fontsCommand, err := typst.NewFontsCommand(
		typstBinary,
		typst.WithoutSystemFonts(),
		typst.WithoutEmbeddedFonts(),
		typst.WithFontPath(filepath.Join(koDataPath, "fonts")),
	)

	if err != nil {
		return err
	}

	indexTemplate, err := template.New("index").Parse(indexHTML)

	if err != nil {
		return fmt.Errorf("could not parse index template: %w", err)
	}

	getRoot := func(w http.ResponseWriter, r *http.Request) {
		fonts, err := fontsCommand.Run()

		if err != nil {
			http.Error(w, fmt.Sprintf("could not list fonts: %s", err.Error()), 500)
			return
		}

		err = indexTemplate.Execute(w, struct{ Fonts []string }{Fonts: fonts})

		if err != nil {
			http.Error(w, fmt.Sprintf("could not render template: %s", err.Error()), 500)
			return
		}
	}

	postRoot := func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()

		if err != nil {
			http.Error(w, fmt.Sprintf("could not parse form: %s", err.Error()), 500)
			return
		}

		body := r.FormValue("source")

		pdf, err := compiler.Run(body)

		if err != nil {
			http.Error(w, fmt.Sprintf("unable to read result PDF: %s", err.Error()), 500)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Write(pdf)
	}

	http.HandleFunc("GET /", getRoot)
	http.HandleFunc("POST /", postRoot)

	return http.ListenAndServe(":8080", nil)
}
