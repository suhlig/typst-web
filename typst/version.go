package typst

import (
	"fmt"
)

func assetSHA(version, name string) (string, error) {
	switch version {
	case "0.14.2":
		switch name {
		case "typst-aarch64-apple-darwin.tar.xz":
			return "sha256:470aa49a2298d20b65c119a10e4ff8808550453e0cb4d85625b89caf0cedf048", nil
		case "typst-x86_64-apple-darwin.tar.xz":
			return "sha256:4e91d8e1e33ab164f949c5762e01ee3faa585c8615a2a6bd5e3677fa8506b249", nil
		case "typst-x86_64-unknown-linux-musl.tar.xz":
			return "sha256:a6044cbad2a954deb921167e257e120ac0a16b20339ec01121194ff9d394996d", nil
		case "typst-aarch64-unknown-linux-musl.tar.xz":
			return "sha256:491b101aa40a3a7ea82a3f8a6232cabb4e6a7e233810082e5ac812d43fdcd47a", nil
		}
	}

	return "", fmt.Errorf("no known SHA for %s %s", version, name)
}

func assetName(version, os, arch string) (string, error) {
	switch version {
	case "0.14.2":
		switch os {
		case "darwin":
			switch arch {
			case "arm64":
				return "typst-aarch64-apple-darwin.tar.xz", nil
			case "amd64":
				return "typst-x86_64-apple-darwin.tar.xz", nil
			}
		case "linux":
			switch arch {
			case "arm64":
				return "typst-aarch64-unknown-linux-musl.tar.xz", nil
			case "amd64":
				return "typst-x86_64-unknown-linux-musl.tar.xz", nil
			}
		}
	}

	return "", fmt.Errorf("unsupported {version, GOOS, GOARCH} combination: {%s, %s, %s}", version, os, arch)
}

// TODO Replace os, arch with Platform
type Platform struct {
	OS   string
	Arch string
}

func SupportedPlatforms() []Platform {
	var result []Platform

	for _, os := range []string{"darwin", "linux"} {
		for _, arch := range []string{"arm64", "amd64"} {
			result = append(result, Platform{OS: os, Arch: arch})
		}
	}

	return result
}
