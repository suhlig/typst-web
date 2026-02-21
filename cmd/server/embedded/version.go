package typstmeta

import (
	"fmt"
	"os"
)

const Version = "0.14.2"

func SHA256(assetName string) string {
	switch assetName {
	case "typst-x86_64-unknown-linux-musl.tar.xz":
		return "sha256:a6044cbad2a954deb921167e257e120ac0a16b20339ec01121194ff9d394996d"
	case "typst-aarch64-apple-darwin.tar.xz":
		return "sha256:470aa49a2298d20b65c119a10e4ff8808550453e0cb4d85625b89caf0cedf048"
	}

	fmt.Fprintf(os.Stderr, "WARN: no known SHA for %s\n", assetName)

	return ""
}
