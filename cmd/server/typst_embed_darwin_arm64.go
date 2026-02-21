 //go:build darwin && arm64

package main

import _ "embed"

//go:embed embedded/typst/darwin/arm64/typst
var typstBin []byte
