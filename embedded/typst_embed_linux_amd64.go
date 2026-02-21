//go:build linux && amd64

package embedded

import _ "embed"

//go:embed typst/linux/amd64/typst
var typstBin []byte
