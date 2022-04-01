package common

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
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

func TestGetYamlNodeInMap(t *testing.T) {
	fakeMap := map[string]string{
		"key1": "thefirstvalue",
		"key2": "thesecondvalue",
		"key3": "thethirdvalue",
	}

	testKey := "key2"

	var fakeNode yaml.Node
	fakeYamlString, err := yaml.Marshal(&fakeMap)
	if err != nil {
		panic(err)
	}

	decoder := yaml.NewDecoder(strings.NewReader(string(fakeYamlString)))

	err = decoder.Decode(&fakeNode)
	if err != nil {
		panic(err)
	}

	returnNode, err := getYamlNodeInMap(fakeNode.Content[0], testKey)

	if err != nil {
		panic(err)
	}

	if returnNode.Value != fakeMap[testKey] {
		t.Errorf("Wrong value for %s: %s", testKey, returnNode.Value)
	}
}
