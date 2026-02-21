 //go:build !(darwin || linux) || !(amd64 || arm64)

package main

var typstBin []byte

func init() {
	panic("no embedded typst binary available for this platform")
}
