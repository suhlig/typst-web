package typstweb

import (
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/suhlig/gosmos/typst"
)

//go:embed index.html
var indexHTML string

var indexTemplate = template.Must(template.New("index").Parse(indexHTML))

// Option is a functional option for Server.
type Option func(*Server)

// WithLogger sets the logger used by the server.
// Without this option, all log output is discarded.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

type Server struct {
	compileCommand *typst.CompileCommand
	fontsCommand   *typst.FontsCommand
	logger         *slog.Logger
}

func NewServer(typstBinPath, fontsPath string, opts ...Option) (*Server, error) {
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

	s := &Server{
		compileCommand: compileCommand,
		fontsCommand:   fontsCommand,
		logger:         slog.New(slog.DiscardHandler),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func (s *Server) Get(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("request", "method", r.Method, "path", r.URL.Path)

	fonts, err := s.fontsCommand.Run()

	if err != nil {
		s.logger.Error("could not list fonts", "err", err)
		http.Error(w, fmt.Sprintf("could not list fonts: %s", err.Error()), 500)
		return
	}

	err = indexTemplate.Execute(w, struct{ Fonts []string }{Fonts: fonts})

	if err != nil {
		s.logger.Error("could not render template", "err", err)
		http.Error(w, fmt.Sprintf("could not render template: %s", err.Error()), 500)
		return
	}
	s.logger.Info("response", "method", r.Method, "path", r.URL.Path, "fontsCount", len(fonts))
}

func (s *Server) Post(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("request", "method", r.Method, "path", r.URL.Path)

	err := r.ParseForm()

	if err != nil {
		s.logger.Error("could not parse form", "err", err)
		http.Error(w, fmt.Sprintf("could not parse form: %s", err.Error()), 500)
		return
	}

	body := r.FormValue("source")

	result, err := s.compileCommand.Run(body)

	if err != nil {
		s.logger.Error("unable to compile typst source", "err", err)
		http.Error(w, fmt.Sprintf("unable to read result PDF: %s", err.Error()), 500)
		return
	}

	for _, d := range result.Diagnostics {
		s.logger.Warn("typst diagnostic", "level", d.Level, "message", d.Message)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Write(result.PDF)
	s.logger.Info("response", "method", r.Method, "path", r.URL.Path, "pdfSize", len(result.PDF))
}
