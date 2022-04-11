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
		t.Log(err)
		t.FailNow()
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
		t.Log(err)
		t.FailNow()
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
		t.Log(err)
		t.FailNow()
	}

	decoder := yaml.NewDecoder(strings.NewReader(string(fakeYamlString)))

	err = decoder.Decode(&fakeNode)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	returnNode, err := getYamlNodeInMap(fakeNode.Content[0], testKey)

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if returnNode.Value != fakeMap[testKey] {
		t.Errorf("Wrong value for %s: %s", testKey, returnNode.Value)
	}
}

func TestGetYamlByReference(t *testing.T) {
	fakeYamlString := `
key1: value1
key2:
  - value2
  - value3
key3:
  nestedKey1: value4
  nestedKey2:
    - value5
    - value6
  nestedKey3: value7
`
	testKey := ".key3.nestedKey2[1]"
	expectedValue := "value6"

	decoder := yaml.NewDecoder(strings.NewReader(string(fakeYamlString)))

	var fakeNode yaml.Node
	err := decoder.Decode(&fakeNode)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	returnNode, err := getYamlByReference(fakeNode.Content[0], testKey)

	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if returnNode.Value != expectedValue {
		t.Errorf("Wrong value for %s: %s", ".key1", returnNode.Value)
	}
}

func TestUpdateYamlFile(t *testing.T) {
	fakeYamlString := `
key1: value1
key2:
  - value2
  - value3
key3:
  nestedKey1: value4
  nestedKey2:
    - value5
    - value6
  nestedKey3: value7
`
	testKey := ".key3.nestedKey2[1]"
	expectedValue := "a new value"

	tmpfile, err := os.CreateTemp("", "")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	t.Log(tmpfile.Name())

	defer os.Remove(tmpfile.Name())

	fmt.Fprint(tmpfile, fakeYamlString)
	UpdateYamlFile(tmpfile.Name(), testKey, expectedValue)

	_, err = tmpfile.Seek(0, 0)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	var fileToTest yaml.Node
	decoder := yaml.NewDecoder(tmpfile)
	err = decoder.Decode(&fileToTest)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	nodeToTest := fileToTest.Content[0].Content[5].Content[3].Content[1]

	if nodeToTest.Value != expectedValue {
		t.Errorf("Wrong value for %s: %s", testKey, nodeToTest.Value)
	}
	err = tmpfile.Close()
	if err != nil {
		t.Error(err)
	}
}
