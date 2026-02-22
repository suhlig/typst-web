package typst

import "fmt"

type Platform struct {
	OS   string
	Arch string
}

func (p Platform) String() string {
	return fmt.Sprintf("{%s, %s}", p.OS, p.Arch)
}

func Platforms(version string) []Platform {
	assets, ok := versions[version]

	if !ok {
		return nil
	}

	seen := make(map[Platform]bool)

	var out []Platform

	for _, a := range assets {
		if !seen[a.Platform] {
			out = append(out, a.Platform)
			seen[a.Platform] = true
		}
	}

	return out
}
