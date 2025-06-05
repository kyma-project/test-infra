package matcher

import "github.com/gobwas/glob"

func Match(pattern, path string) (bool, error) {
	g, err := glob.Compile(pattern)
	if err != nil {
		return false, err
	}

	return g.Match(path), nil
}
