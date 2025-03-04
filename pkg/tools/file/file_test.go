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

func TestFindAllRecursively(t *testing.T) {
	//given
	tmpDir, err := ioutil.TempDir("", "TestFindAllRecursively")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	expectedPaths, err := createFiles(tmpDir)
	require.NoError(t, err)

	var testCases = []struct {
		name       string
		rootPath   string
		extension  string
		result     []string
		shouldFail bool
	}{
		{"Valid path", tmpDir, ".yaml", expectedPaths, false},
		{"No files", tmpDir, ".nope", []string{}, false},
		{"Invalid root path", fmt.Sprintf("%s/%s/trap", tmpDir, dirStructure), "", nil, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			result, err := FindAllRecursively(testCase.rootPath, testCase.extension)

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

func createFiles(root string) ([]string, error) {
	err := os.MkdirAll(fmt.Sprintf("%s/%s", root, dirStructure), os.ModePerm)
	if err != nil {
		return nil, err
	}

	var filePaths []string
	for _, path := range files {
		filePath := fmt.Sprintf("%s/%s", root, path)
		err := createFile(filePath)
		if err != nil {
			return nil, err
		}

		filePaths = append(filePaths, filePath)
	}

	return filePaths, nil
}

func createFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}
# (2025-03-04)