package sets

import "strings"

type Strings []string

func (t *Strings) String() string {
	return strings.Join(*t, ",")
}

func (t *Strings) Set(val string) error {
	*t = append(*t, val)
	return nil
}
# (2025-03-04)