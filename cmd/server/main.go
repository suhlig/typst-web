package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"typstweb/typstweb"

	"github.com/suhlig/gosmos/typst"
)

//go:generate go tool github.com/suhlig/gosmos/cmd/fetch-typst --path kodata/typst
//go:generate go tool github.com/suhlig/gosmos/cmd/fetch-fonts --path kodata/fonts "Meie Script" Roboto"

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

	typstBinPath := typst.CurrentBinPath(filepath.Join(koDataPath, "typst"))
	fontsPath := filepath.Join(koDataPath, "fonts")

	srv, err := typstweb.NewServer(typstBinPath, fontsPath)

	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", srv.Get)
	mux.HandleFunc("POST /", srv.Post)

	return http.ListenAndServe(":8080", mux)
}
