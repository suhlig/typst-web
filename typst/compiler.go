package typst

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Compiler struct {
	binPath   string
	fontPaths []string
}

func (t Compiler) Help() (string, error) {
	cmd := exec.Command(t.binPath, "help")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return stdout.String(), nil
}

func NewCompiler(binPath string) (*Compiler, error) {
	t := Compiler{binPath: binPath}

	_, err := t.Help()

	if err != nil {
		return nil, err
	}

	return &t, nil
}

// TODO TYPST_ROOT
// TODO TYPST_IGNORE_SYSTEM_FONTS
// TODO TYPST_IGNORE_EMBEDDED_FONTS
// TODO TYPST_PACKAGE_PATH
// TODO TYPST_PACKAGE_CACHE_PATH
// TODO SOURCE_DATE_EPOCH
// TODO TYPST_FEATURES

func (c *Compiler) WithFontPath(p string) *Compiler {
	c.fontPaths = append(c.fontPaths, p)
	return c
}

func (t Compiler) Compile(input string) ([]byte, error) {
	pdfPath, err := os.CreateTemp("", "out.pdf")

	cmd := exec.Command(
		t.binPath,
		"compile",
		"-",
		"--ignore-system-fonts",
		"--ignore-embedded-fonts",
		"--format", "pdf",
		pdfPath.Name(),
	)

	if len(t.fontPaths) > 0 {
		cmd.Args = append(cmd.Args, "--font-path", strings.Join(t.fontPaths, ":"))
	}

	stdin, err := cmd.StdinPipe()

	if err != nil {
		return nil, fmt.Errorf("unable to get stdin pipe: %s", err.Error())
	}

	// TODO Accept readers or channels where stdout and stderr can be piped to
	// so that clients get to present errors and warnings
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	go func() {
		stdin.Write([]byte(input))
		stdin.Close()
	}()

	err = cmd.Run()

	fmt.Fprintf(os.Stderr, "STDOUT:\n%s\n", stdout.String())
	fmt.Fprintf(os.Stderr, "STDERR:\n%s\n", stderr.String())

	if err != nil {
		return nil, fmt.Errorf("failed to run command %s: %w", cmd, err)
	}

	pdf, err := os.ReadFile(pdfPath.Name())

	if err != nil {
		return nil, fmt.Errorf("unable to read result PDF: %w", err)
	}

	defer os.Remove(pdfPath.Name())

	return pdf, nil
}
