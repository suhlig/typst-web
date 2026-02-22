package typst

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Platform struct {
	OS   string
	Arch string
}

func (p Platform) String() string {
	return fmt.Sprintf("{%s, %s}", p.OS, p.Arch)
}

// assetInfo is the single source of truth for each asset: name, sha, and platform.
type assetInfo struct {
	Name     string
	SHA      string
	Platform Platform
}

// versions holds all versioned assets. Treat this as read-only.
var versions = map[string][]assetInfo{
	"0.14.2": {
		{
			Name: "typst-aarch64-apple-darwin.tar.xz",
			SHA:  "sha256:470aa49a2298d20b65c119a10e4ff8808550453e0cb4d85625b89caf0cedf048",
			Platform: Platform{
				OS:   "darwin",
				Arch: "arm64",
			},
		},
		{
			Name: "typst-x86_64-apple-darwin.tar.xz",
			SHA:  "sha256:4e91d8e1e33ab164f949c5762e01ee3faa585c8615a2a6bd5e3677fa8506b249",
			Platform: Platform{
				OS:   "darwin",
				Arch: "amd64",
			},
		},
		{
			Name: "typst-x86_64-unknown-linux-musl.tar.xz",
			SHA:  "sha256:a6044cbad2a954deb921167e257e120ac0a16b20339ec01121194ff9d394996d",
			Platform: Platform{
				OS:   "linux",
				Arch: "amd64",
			},
		},
		{
			Name: "typst-aarch64-unknown-linux-musl.tar.xz",
			SHA:  "sha256:491b101aa40a3a7ea82a3f8a6232cabb4e6a7e233810082e5ac812d43fdcd47a",
			Platform: Platform{
				OS:   "linux",
				Arch: "arm64",
			},
		},
	},
}

func assetSHA(version, name string) (string, error) {
	assets, ok := versions[version]
	if !ok {
		return "", fmt.Errorf("no known version %s", version)
	}
	for _, a := range assets {
		if a.Name == name {
			return a.SHA, nil
		}
	}
	return "", fmt.Errorf("no known SHA for %s %s", version, name)
}

func assetName(version, os, arch string) (string, error) {
	assets, ok := versions[version]
	if !ok {
		return "", fmt.Errorf("unsupported {version, GOOS, GOARCH} combination: {%s, %s, %s}", version, os, arch)
	}
	for _, a := range assets {
		if a.Platform.OS == os && a.Platform.Arch == arch {
			return a.Name, nil
		}
	}
	return "", fmt.Errorf("unsupported {version, GOOS, GOARCH} combination: {%s, %s, %s}", version, os, arch)
}

func Platforms() []Platform {
	return platformsForVersion(latestVersion())
}

func platformsForVersion(version string) []Platform {
	assets, ok := versions[version]
	if !ok {
		return nil
	}
	seen := make(map[string]bool)
	var out []Platform
	for _, a := range assets {
		key := a.Platform.OS + "/" + a.Platform.Arch
		if !seen[key] {
			out = append(out, a.Platform)
			seen[key] = true
		}
	}
	return out
}

func latestVersion() string {
	if len(versions) == 0 {
		return ""
	}
	keys := make([]string, 0, len(versions))
	for k := range versions {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		mi, ni, pi := semverParts(keys[i])
		mj, nj, pj := semverParts(keys[j])
		if mi != mj {
			return mi < mj
		}
		if ni != nj {
			return ni < nj
		}
		return pi < pj
	})
	return keys[len(keys)-1]
}

func semverParts(s string) (int, int, int) {
	parts := strings.SplitN(s, ".", 3)
	var m, n, p int
	if len(parts) > 0 {
		m, _ = strconv.Atoi(parts[0])
	}
	if len(parts) > 1 {
		n, _ = strconv.Atoi(parts[1])
	}
	if len(parts) > 2 {
		ps := parts[2]
		if idx := strings.IndexAny(ps, "-+"); idx >= 0 {
			ps = ps[:idx]
		}
		p, _ = strconv.Atoi(ps)
	}
	return m, n, p
}
