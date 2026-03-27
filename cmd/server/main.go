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

	typstBinPath := filepath.Join(koDataPath, "typst", runtime.GOOS, runtime.GOARCH, "typst")
	fontsPath := filepath.Join(koDataPath, "fonts")

	compileCommand, err := typst.NewCompileCommand(
		typstBinPath,
		typst.WithoutSystemFonts(),
		typst.WithoutEmbeddedFonts(),
		typst.WithFontPath(fontsPath),
	)

	if err != nil {
		return err
	}

	fontsCommand, err := typst.NewFontsCommand(
		typstBinPath,
		typst.WithoutSystemFonts(),
		typst.WithoutEmbeddedFonts(),
		typst.WithFontPath(fontsPath),
	)

	if err != nil {
		return err
	}

	indexTemplate, err := template.New("index").Parse(indexHTML)

	if err != nil {
		return fmt.Errorf("could not parse index template: %w", err)
	}

	srv := &typstweb.Server{
		CompileCommand: compileCommand,
		FontsCommand:   fontsCommand,
		Template:       indexTemplate,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", srv.Get)
	mux.HandleFunc("POST /", srv.Post)

	return http.ListenAndServe(":8080", mux)
}
