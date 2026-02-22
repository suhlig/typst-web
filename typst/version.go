package typst

import (
	"fmt"
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
	var result []Platform
	for _, os := range []string{"darwin", "linux"} {
		for _, arch := range []string{"arm64", "amd64"} {
			result = append(result, Platform{OS: os, Arch: arch})
		}
	}
	return result
}
