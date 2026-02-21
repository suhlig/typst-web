//go:build darwin && amd64

package embedded

import _ "embed"

//go:embed typst/darwin/amd64/typst
var typstBin []byte
