package typst

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type Typst struct {
	binPath string
}

func (t Typst) Help() (string, error) {
	cmd := exec.Command(t.binPath, "help")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return stdout.String(), nil
}

func NewTypst(binPath string) (*Typst, error) {
	t := Typst{binPath: binPath}

	_, err := t.Help()

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (t Typst) Compile(input string) ([]byte, error) {
	pdfPath, err := os.CreateTemp("", "out.pdf")

	cmd := exec.Command(t.binPath, "compile", "-", "--format", "pdf", pdfPath.Name())

	stdin, err := cmd.StdinPipe()

	if err != nil {
		return nil, fmt.Errorf("unable to get stdin pipe: %s", err.Error())
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	go func() {
		stdin.Write([]byte(input))
		stdin.Close()
	}()

	err = cmd.Run()

	if err != nil {
		// fmt.Fprintf(os.Stderr, "STDOUT:\n%s\n", stdout.String())
		// fmt.Fprintf(os.Stderr, "STDERR:\n%s\n", stderr.String())
		return nil, fmt.Errorf("failed to run command %s: %w", cmd, err)
	}

	pdf, err := os.ReadFile(pdfPath.Name())

	if err != nil {
		return nil, fmt.Errorf("unable to read result PDF: %w", err)
	}

	defer os.Remove(pdfPath.Name())

	return pdf, nil
}
