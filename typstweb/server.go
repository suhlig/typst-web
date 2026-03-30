package typstweb

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/suhlig/gosmos/typst"
)

//go:embed index.html
var indexHTML string

var indexTemplate = template.Must(template.New("index").Parse(indexHTML))

type Server struct {
	compileCommand *typst.CompileCommand
	fontsCommand   *typst.FontsCommand
}

func NewServer(typstBinPath, fontsPath string) (*Server, error) {
	compileCommand, err := typst.NewCompileCommand(
		typstBinPath,
		typst.WithoutSystemFonts(),
		typst.WithoutEmbeddedFonts(),
		typst.WithFontPath(fontsPath),
	)

	if err != nil {
		return nil, err
	}

	fontsCommand, err := typst.NewFontsCommand(
		typstBinPath,
		typst.WithoutSystemFonts(),
		typst.WithoutEmbeddedFonts(),
		typst.WithFontPath(fontsPath),
	)

	if err != nil {
		return nil, err
	}

	return &Server{
		compileCommand: compileCommand,
		fontsCommand:   fontsCommand,
	}, nil
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
	fonts, err := s.fontsCommand.Run()

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

func (s *Server) Post(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse form: %s", err.Error()), 500)
		return
	}

	body := r.FormValue("source")

	result, err := s.compileCommand.Run(body)

	if err != nil {
		http.Error(w, fmt.Sprintf("unable to read result PDF: %s", err.Error()), 500)
		return
	}

	for _, d := range result.Diagnostics {
		fmt.Printf("[%s] %s\n", d.Level, d.Message)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Write(result.PDF)
}
