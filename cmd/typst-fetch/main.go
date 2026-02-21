package main

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	typstmeta "typstweb/cmd/server/embedded"

	"github.com/google/go-github/v83/github"
	"github.com/ulikunitz/xz"
	"golang.org/x/oauth2"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	client := newGHClient(ctx)
	const owner, repo = "typst", "typst"

	// Try v<version>, then <version>
	tag := "v" + typstmeta.Version

	rel, resp, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)

	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			tag = typstmeta.Version
			rel, resp, err = client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
		}
	}

	if err != nil {
		return fmt.Errorf("get release by tag %q: %w", tag, err)
	}

	asset, err := pickAsset(rel.Assets)

	if err != nil {
		return err
	}

	log.Printf("selected asset: %s (%d bytes)", asset.GetName(), asset.GetSize())

	data, err := download(ctx, client, owner, repo, asset)

	if err != nil {
		return fmt.Errorf("download asset %q: %w", asset.GetName(), err)
	}

	log.Printf("downloaded %d bytes", len(data))

	// Verify checksum only for the tar.xz asset; the SHA tracks the archive, not the extracted binary.
	err = verifySHA(data, typstmeta.SHA256(*asset.Name))

	if err != nil {
		return err
	}

	log.Printf("sha256 verified")

	bin, err := extractTypstBinary(data)

	if err != nil {
		return err
	}

	if len(bin) == 0 {
		return errors.New("extracted typst binary is empty")
	}

	outPath := filepath.FromSlash(filepath.Join("typst", runtime.GOOS, runtime.GOARCH, "typst"))

	if err := atomicWrite(outPath, bin, 0644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}

	log.Printf("wrote %d bytes to %s", len(bin), outPath)

	return nil
}

func newGHClient(ctx context.Context) *github.Client {
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})
		tc := oauth2.NewClient(ctx, ts)
		return github.NewClient(tc)
	}
	return github.NewClient(nil)
}

func pickAsset(assets []*github.ReleaseAsset) (*github.ReleaseAsset, error) {
	var preferred string

	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "arm64":
			preferred = "typst-aarch64-apple-darwin.tar.xz"
		case "amd64":
			preferred = "typst-x86_64-apple-darwin.tar.xz"
		default:
			return nil, fmt.Errorf("unsupported {GOOS, GOARCH} combination: {%s, %s}", runtime.GOOS, runtime.GOARCH)
		}
	case "linux":
		switch runtime.GOARCH {
		case "arm64":
			preferred = " typst-aarch64-unknown-linux-musl.tar.xz "
		case "amd64":
			preferred = "typst-x86_64-unknown-linux-musl.tar.xz"
		default:
			return nil, fmt.Errorf("unsupported {GOOS, GOARCH} combination: {%s, %s}", runtime.GOOS, runtime.GOARCH)
		}
	}

	for _, a := range assets {
		if a.GetName() == preferred {
			return a, nil
		}
	}

	return nil, errors.New("no suitable release asset found")
}

func assetNames(assets []*github.ReleaseAsset) string {
	var ss []string
	for _, a := range assets {
		ss = append(ss, a.GetName())
	}
	return strings.Join(ss, ", ")
}

func download(ctx context.Context, client *github.Client, owner, repo string, asset *github.ReleaseAsset) ([]byte, error) {
	// Prefer the browser download URL when available.
	if u := asset.GetBrowserDownloadURL(); u != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("download %s: status %s", asset.GetName(), resp.Status)
		}
		return io.ReadAll(resp.Body)
	}

	// Fallback to API streaming if needed.
	rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, asset.GetID(), http.DefaultClient)
	if err != nil {
		return nil, err
	}
	if rc == nil {
		return nil, errors.New("no stream for asset")
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func verifySHA(data []byte, expect string) error {
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])

	exp := strings.TrimSpace(expect)
	if strings.HasPrefix(exp, "sha256:") {
		exp = strings.TrimPrefix(exp, "sha256:")
	}
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
			return nil, fmt.Errorf("tar read: %w", err)
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		// Look for the binary named "typst" (may be inside a directory).
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

func atomicWrite(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	// Best-effort sync before rename.
	if f, err := os.Open(tmp); err == nil {
		_ = f.Sync()
		_ = f.Close()
	}
	return os.Rename(tmp, path)
}
