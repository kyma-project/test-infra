package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dirStructure = "it/is/a"
)

var files = []string{"it/fileA.yaml", "it/is/fileB.yaml", "it/is/a/fileC.yaml"}

func TestFindAllRec(t *testing.T) {
	//given
	tmpDir, err := ioutil.TempDir("", "TestFindAllRec")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	err = createFiles(tmpDir)
	require.NoError(t, err)

	var testCases = []struct {
		name       string
		rootPath   string
		extension  string
		result     []string
		shouldFail bool
	}{
		{"Valid path", tmpDir, ".yaml", files, false},
		{"No files", tmpDir, ".nope", []string{}, false},
		{"Invalid root path", fmt.Sprintf("%s/%s/trap", tmpDir, dirStructure), "", nil, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			result, err := FindAllRec(testCase.rootPath, testCase.extension)

			//then
			if testCase.shouldFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, len(testCase.result))
				assert.ElementsMatch(t, testCase.result, result)
			}
		})
	}
}

func createFiles(root string) error {
	err := os.MkdirAll(fmt.Sprintf("%s/%s", root, dirStructure), os.ModePerm)
	if err != nil {
		return err
	}

	for _, path := range files {
		err := createFile(fmt.Sprintf("%s/%s", root, path))
		if err != nil {
			return err
		}
	}

	return nil
}

func createFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}
