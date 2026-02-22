package typst

// assetInfo is the single source of truth for each asset: name, sha, and platform.
type assetInfo struct {
	Platform Platform
	Name     string
	SHA      string
}
