package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"typstweb/typst"

	"golang.org/x/sync/errgroup"
)

func main() {
	err := mainE(context.Background())

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s\n", err)
		os.Exit(1)
	}
}

const typstVersion = "0.14.2"

func mainE(ctx context.Context) error {
	fetcher := typst.NewFetcher(typstVersion)
	g, ctx := errgroup.WithContext(ctx)

	for _, platform := range typst.Platforms(typstVersion) {
		p := platform // capture

		g.Go(func() error {
			bin, err := fetcher.Fetch(ctx, p)

			if err != nil {
				return err
			}

			path := filepath.FromSlash(filepath.Join("kodata", "typst", p.OS, p.Arch, "typst"))

			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			if err := os.WriteFile(path, bin, 0755); err != nil {
				return fmt.Errorf("writing %s: %w", path, err)
			}

			return nil
		})
	}

	return g.Wait()
}
