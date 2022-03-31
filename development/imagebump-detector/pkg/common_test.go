package common

import (
	"fmt"
	"os"
	"testing"
)

func TestParseNotationFile(t *testing.T) {
	fakeKey := ".key1.key2.key3"
	fakePath := "filepath/thatis/right"

	tmpfile, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}

	defer os.Remove(tmpfile.Name())

	fmt.Fprintf(tmpfile, `
normal content file
# this line is ignored because: it doesn't match the pattern
# %s:%s
whatever additional content
# otherfilepath/thatdoesntwork:.otherkey
`, fakePath, fakeKey)

	parsedPath, parsedKey, err := ParseNotationFile(tmpfile.Name())

	if err != nil {
		panic(err)
	}

	if parsedPath != fakePath {
		t.Errorf("Wrong filepath: %s", parsedPath)
	}

	if parsedKey != fakeKey {
		t.Errorf("Wrong key: %s", parsedKey)
	}
}
