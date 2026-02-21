 //go:build linux && arm64

package main

import _ "embed"

//go:embed embedded/typst/linux/arm64/typst
var typstBin []byte
