package typst

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v83/github"
	"github.com/ulikunitz/xz"
)

type Fetcher struct {
	version  string
	ghClient *github.Client
}

func NewFetcher(version string) *Fetcher {
	client := github.NewClient(nil)

	return &Fetcher{
		version:  version,
		ghClient: client,
	}
}

func (f Fetcher) Fetch(ctx context.Context, platform Platform) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	tag := "v" + f.version

	rel, _, err := f.ghClient.Repositories.GetReleaseByTag(ctx, "typst", "typst", tag)

	if err != nil {
		return nil, fmt.Errorf("getting the release by tag %q: %w", tag, err)
	}

	asset, err := f.pickAsset(rel.Assets, platform)

	if err != nil {
		return nil, err
	}

	data, err := f.download(ctx, asset.GetBrowserDownloadURL())

	if err != nil {
		return nil, fmt.Errorf("downloading asset %q: %w", asset.GetName(), err)
	}

	sha, err := assetSHA(f.version, *asset.Name)

	if err != nil {
		return nil, err
	}

	err = verifySHA(data, sha)

	if err != nil {
		return nil, err
	}

	return extractTypstBinary(data)
}

func (f Fetcher) pickAsset(assets []*github.ReleaseAsset, platform Platform) (*github.ReleaseAsset, error) {
	assetName, err := assetName(f.version, platform.OS, platform.Arch)

	if err != nil {
		return nil, err
	}

	for _, a := range assets {
		if a.GetName() == assetName {
			return a, nil
		}
	}

	return nil, fmt.Errorf("no asset found for %s", platform)
}

func (f Fetcher) download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: status %s", url, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func verifySHA(data []byte, expect string) error {
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])

	exp := strings.TrimSpace(expect)
	exp = strings.TrimPrefix(exp, "sha256:")
	exp = strings.ToLower(exp)

	if got != exp {
		return fmt.Errorf("sha256 mismatch: got %s, expected %s", got, exp)
	}

	return nil
}

func extractTypstBinary(data []byte) ([]byte, error) {
	xr, err := xz.NewReader(bytes.NewReader(data))

	if err != nil {
		return nil, fmt.Errorf("xz reader: %w", err)
	}

	tr := tar.NewReader(xr)

	for {
		hdr, err := tr.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("reading tar: %w", err)
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		if filepath.Base(hdr.Name) == "typst" {
			b, err := io.ReadAll(tr)

			if err != nil {
				return nil, err
			}

			return b, nil
		}
	}

	return nil, errors.New("typst binary not found in archive")
}
