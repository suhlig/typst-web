package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"typstweb/typstweb"

	"github.com/suhlig/gosmos/typst"
)

//go:generate go tool github.com/suhlig/gosmos/cmd/fetch-typst --path kodata/typst
//go:generate go tool github.com/suhlig/gosmos/cmd/fetch-fonts --path kodata/fonts "Meie Script" Roboto

func main() {
	err := mainE()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s\n", err)
		os.Exit(1)
	}
}

func mainE() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	koDataPath := os.Getenv("KO_DATA_PATH")
	logger.Info("using", "KO_DATA_PATH", koDataPath)

	typstBinPath := typst.CurrentBinPath(filepath.Join(koDataPath, "typst"))
	logger.Info("using typst binary from", "typstBinPath", typstBinPath)

	fontsPath := filepath.Join(koDataPath, "fonts")
	logger.Info("reading fonts from", "fontsPath", fontsPath)

	server, err := typstweb.NewServer(typstBinPath, fontsPath, typstweb.WithLogger(logger.WithGroup("server")))

	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", server.Get)
	mux.HandleFunc("POST /", server.Post)

	return http.ListenAndServe(":8080", mux)
}
