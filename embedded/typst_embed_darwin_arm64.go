//go:build darwin && arm64

package embedded

import _ "embed"

//go:embed typst/darwin/arm64/typst
var typstBin []byte
