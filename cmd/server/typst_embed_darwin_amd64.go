//go:build darwin && amd64

package main

import _ "embed"

//go:embed cmd/server/embedded/typst/darwin/amd64/typst
var typstBin []byte
