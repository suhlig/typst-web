package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"typstweb/typst"
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

	platforms := typst.Platforms(typstVersion)
	if len(platforms) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for _, platform := range platforms {
		platform := platform // capture loop var
		wg.Add(1)
		go func() {
			defer wg.Done()

			bin, err := fetcher.Fetch(ctx, platform)
			if err != nil {
				select {
				case errCh <- err:
					cancel()
				default:
				}
				return
			}

			path := filepath.FromSlash(filepath.Join("kodata", "typst", platform.OS, platform.Arch, "typst"))

			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				select {
				case errCh <- err:
					cancel()
				default:
				}
				return
			}

			if err := os.WriteFile(path, bin, 0755); err != nil {
				select {
				case errCh <- fmt.Errorf("writing %s: %w", path, err):
					cancel()
				default:
				}
				return
			}
		}()
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
