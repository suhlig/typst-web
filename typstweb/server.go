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
	CompileCommand *typst.CompileCommand
	FontsCommand   *typst.FontsCommand
}

func NewServer(compileCommand *typst.CompileCommand, fontsCommand *typst.FontsCommand) *Server {
	return &Server{
		CompileCommand: compileCommand,
		FontsCommand:   fontsCommand,
	}
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
	fonts, err := s.FontsCommand.Run()

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

	pdf, err := s.CompileCommand.Run(body)

	if err != nil {
		http.Error(w, fmt.Sprintf("unable to read result PDF: %s", err.Error()), 500)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Write(pdf)
}
