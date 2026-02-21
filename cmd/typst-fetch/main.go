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
	"strings"
	"time"

	"github.com/google/go-github/v83/github"
	"github.com/ulikunitz/xz"
	"golang.org/x/oauth2"

	"typstweb/internal/typstmeta"
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

	if err := verifySHA(data, typstmeta.SHA256); err != nil {
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

	outPath := filepath.FromSlash("internal/embedded/typst/typst")
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
	// Prefer Linux x86_64 musl tar.xz as per project README and defaults.
	const preferred = "x86_64-unknown-linux-musl.tar.xz"

	var candidate *github.ReleaseAsset
	for _, a := range assets {
		name := a.GetName()
		if strings.Contains(name, preferred) {
			return a, nil
		}
		// Reasonable fallbacks: any linux x86_64 tar.xz, or a plain binary named typst.
		if candidate == nil && strings.HasSuffix(name, ".tar.xz") && strings.Contains(name, "linux") && strings.Contains(name, "x86_64") {
			candidate = a
		}
		if candidate == nil && (strings.EqualFold(name, "typst") || strings.HasSuffix(name, "typst-linux-x86_64")) {
			candidate = a
		}
	}
	if candidate != nil {
		return candidate, nil
	}
	return nil, fmt.Errorf("no suitable release asset found; available: %s", assetNames(assets))
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
	if !isXZ(data) {
		// Assume the asset is the raw binary.
		return data, nil
	}
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

func isXZ(b []byte) bool {
	if len(b) < 6 {
		return false
	}
	// xz magic: FD 37 7A 58 5A 00
	return b[0] == 0xFD && b[1] == 0x37 && b[2] == 0x7A && b[3] == 0x58 && b[4] == 0x5A && b[5] == 0x00
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
