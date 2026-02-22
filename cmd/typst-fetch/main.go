package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"typstweb/typst"
)

func main() {
	err := mainE(context.Background())

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s\n", err)
		os.Exit(1)
	}
}

func mainE(ctx context.Context) error {
	fetcher := typst.NewFetcher("0.14.2")

	// TODO in parallel
	for _, platform := range typst.Platforms() {
		bin, err := fetcher.Fetch(ctx, platform)

		if err != nil {
			return err
		}

		path := filepath.FromSlash(filepath.Join("kodata", "typst", platform.OS, platform.Arch, "typst"))

		err = os.MkdirAll(filepath.Dir(path), 0755)

		if err != nil {
			return err
		}

		err = os.WriteFile(path, bin, 0755)

		if err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}

	return nil
}
