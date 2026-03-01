package main

import (
	"context"
	"errors"
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
		g.Go(func() error {
			path := filepath.FromSlash(filepath.Join("kodata", "typst", platform.OS, platform.Arch, "typst"))

			fileExists, err := fileExists(path)

			if err != nil {
				return err
			}

			if fileExists {
				// TODO do we need to check more than the file's existence? Size > 0; maybe even a digest?
				return nil
			}

			bin, err := fetcher.Fetch(ctx, platform)

			if err != nil {
				return err
			}

			err = os.MkdirAll(filepath.Dir(path), 0755)

			if err != nil {
				return err
			}

			err = os.WriteFile(path, bin, 0755)

			if err != nil {
				return fmt.Errorf("writing %s: %w", path, err)
			}

			return nil
		})
	}

	return g.Wait()
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err // permission or other error
}
