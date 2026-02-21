//go:build linux && amd64

package main

import _ "embed"

//go:embed cmd/server/embedded/typst/linux/amd64/typst
var typstBin []byte
