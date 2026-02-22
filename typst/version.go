package typst

var versions = map[string][]assetInfo{
	"0.14.2": {
		{
			Platform: Platform{OS: "darwin", Arch: "arm64"},
			Name:     "typst-aarch64-apple-darwin.tar.xz",
			SHA:      "sha256:470aa49a2298d20b65c119a10e4ff8808550453e0cb4d85625b89caf0cedf048",
		},
		{
			Platform: Platform{OS: "darwin", Arch: "amd64"},
			Name:     "typst-x86_64-apple-darwin.tar.xz",
			SHA:      "sha256:4e91d8e1e33ab164f949c5762e01ee3faa585c8615a2a6bd5e3677fa8506b249",
		},
		{
			Platform: Platform{OS: "linux", Arch: "amd64"},
			Name:     "typst-x86_64-unknown-linux-musl.tar.xz",
			SHA:      "sha256:a6044cbad2a954deb921167e257e120ac0a16b20339ec01121194ff9d394996d",
		},
		{
			Platform: Platform{OS: "linux", Arch: "arm64"},
			Name:     "typst-aarch64-unknown-linux-musl.tar.xz",
			SHA:      "sha256:491b101aa40a3a7ea82a3f8a6232cabb4e6a7e233810082e5ac812d43fdcd47a",
		},
	},
}
