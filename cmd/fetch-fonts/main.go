package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"typstweb/fontsource"

	"golang.org/x/sync/errgroup"
)

func main() {
	err := mainE(context.Background(), os.Args[1:]...)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %s\n", err)
		os.Exit(1)
	}
}

func mainE(ctx context.Context, args ...string) error {
	var allTasks []fontsource.DownloadTask

	for _, family := range args {
		tasks, err := fontsource.GetDownloadTasks(ctx, family)

		if err != nil {
			return fmt.Errorf("getting tasks for %s: %w", family, err)
		}

		allTasks = append(allTasks, tasks...)
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10) // Limit concurrency to avoid hitting rate limits or OS file descriptor limits

	for _, task := range allTasks {
		g.Go(func() error {
			exists, err := fileExists(task.Path)

			if err != nil {
				return err
			}

			if exists {
				return nil
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, task.URL, nil)

			if err != nil {
				return err
			}

			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				return err
			}

			defer resp.Body.Close()

			err = os.MkdirAll(filepath.Dir(task.Path), 0755)

			if err != nil {
				return err
			}

			out, err := os.Create(task.Path)

			if err != nil {
				return err
			}

			defer out.Close()

			_, err = io.Copy(out, resp.Body)

			return err
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
