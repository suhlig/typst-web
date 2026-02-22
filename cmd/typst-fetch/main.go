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
	const typstVersion = "0.14.2"
	fetcher := typst.NewFetcher()

	for _, platform := range supportedPlatforms() {
		bin, err := fetcher.Fetch(ctx, typstVersion, platform.OS, platform.Arch)

		if err != nil {
			return err
		}

		path := filepath.FromSlash(filepath.Join("kodata", "typst", typstVersion, platform.OS, platform.Arch, "typst"))

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

type Platform struct {
	OS   string
	Arch string
}

func supportedPlatforms() []Platform {
	var result []Platform

	for _, os := range []string{"darwin", "linux"} {
		for _, arch := range []string{"arm64", "amd64"} {
			result = append(result, Platform{OS: os, Arch: arch})
		}
	}

	return result
}
