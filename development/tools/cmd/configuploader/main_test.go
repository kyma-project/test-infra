package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"path/filepath"
)

const (
	dirStructure = "where/is/my/test"
)

var files = [...]string{"where/fileA.yaml", "where/is/fileB.yaml", "where/is/my/fileC.yaml"}

func TestReplaceConfigMapFromFile(t *testing.T) {
	//given
	tmpDir, err := ioutil.TempDir("", "TestReplaceConfigMapFromFile")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	file := fmt.Sprintf("%s/fileA.yaml", tmpDir)
	err = createFile(file, "test data")
	require.NoError(t, err)

	client := fake.NewSimpleClientset()
	client.CoreV1().ConfigMaps("default").Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	})

	var testCases = []struct {
		name       string
		path       string
		data       map[string]string
		shouldFail bool
	}{
		{"Valid path", file, map[string]string{"test.yaml": "test data"}, false},
		{"Not existing file", fmt.Sprintf("%s/trap", tmpDir), nil, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			err := replaceConfigMapFromFile("test", testCase.path, client.CoreV1().ConfigMaps("default"))

			//then
			if testCase.shouldFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				config, err := client.CoreV1().ConfigMaps("default").Get("test", metav1.GetOptions{})
				require.NoError(t, err)
				require.NotNil(t, config)
				assert.Len(t, config.Data, 1)
				assert.Equal(t, testCase.data, config.Data)
			}
		})
	}
}

func TestReplaceConfigMapFromDirectory(t *testing.T) {
	//given
	tmpDir, err := ioutil.TempDir("", "TestReplaceConfigMapFromDirectory")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = createFiles(tmpDir)
	require.NoError(t, err)

	client := fake.NewSimpleClientset()
	client.CoreV1().ConfigMaps("default").Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	})

	var testCases = []struct {
		name         string
		path         string
		entriesCount int
		shouldFail   bool
	}{
		{"Multiple files", tmpDir, len(files), false},
		{"No files", fmt.Sprintf("%s/%s", tmpDir, dirStructure), 0, false},
		{"Not existing root", "nope/nope/nope", 0, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//when
			err := replaceConfigMapFromDirectory("test", testCase.path, client.CoreV1().ConfigMaps("default"))

			//then
			if testCase.shouldFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				config, err := client.CoreV1().ConfigMaps("default").Get("test", metav1.GetOptions{})
				require.NoError(t, err)
				require.NotNil(t, config)
				assert.Len(t, config.Data, testCase.entriesCount)
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
		data := fmt.Sprintf("test data %s", filepath.Base(path))
		err := createFile(fmt.Sprintf("%s/%s", root, path), data)
		if err != nil {
			return err
		}
	}

	return nil
}

func createFile(path, data string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(data)

	return err
}
