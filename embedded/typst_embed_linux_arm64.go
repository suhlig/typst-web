//go:build linux && arm64

package embedded

import _ "embed"

//go:embed typst/linux/arm64/typst
var typstBin []byte
