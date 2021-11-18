package promote

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
	"gopkg.in/yaml.v3"
)

func GetWalkFunc(ResourcesDirectoryClean, targetContainerRegistry, targetTag string, dryRun bool, images *[]list.Image) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		//pass the error further, this shouldn't ever happen
		if err != nil {
			return errors.New(err.Error() + " in file " + path)
		}

		// skip directory entries, we just want files
		if info.IsDir() {
			return nil
		}

		// we only want to check values.yaml files
		if info.Name() != "values.yaml" {
			return nil
		}

		var parsedFile yaml.Node
		var parsedImagesFile list.ValueFile
		lines := make([]string, 0)

		yamlFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error while opening %s file: %s", path, err)
		}
		defer yamlFile.Close()

		decoder := yaml.NewDecoder(yamlFile)
		err = decoder.Decode(&parsedFile)
		if err != nil {
			return fmt.Errorf("error while unmarshalling %s file: %s", path, err)
		}

		// rewind and load the file into array of lines
		yamlFile.Seek(0, 0)
		scanner := bufio.NewScanner(yamlFile)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if scanner.Err() != nil {
			return fmt.Errorf("error while reading %s file: %s", path, scanner.Err())
		}

		// get list of images
		yamlFile.Seek(0, 0)
		decoder = yaml.NewDecoder(yamlFile)
		err = decoder.Decode(&parsedImagesFile)
		if err != nil {
			return fmt.Errorf("error while decoding %s file: %s", path, err)
		}
		appendImagesToList(parsedImagesFile, images)

		globalNode := getYamlNode(parsedFile.Content[0], "global")
		if globalNode == nil {
			// skip the whole file
			return nil
		}

		// promote container registry
		if targetContainerRegistry != "" {
			containerRegistryNode := getYamlNode(globalNode, "containerRegistry")
			if containerRegistryNode == nil {
				// skip files without images
				return nil
			}

			containerRegistryPathNode := getYamlNode(containerRegistryNode, "path")
			if containerRegistryPathNode == nil {
				// raise error if the containerRegistry is defined, but psth is not
				return fmt.Errorf("error in %s file: could not find global.containerRegistry.path key", path)
			}
			containerRegistryPathNode.Value = targetContainerRegistry

			outputLine, err := yamlNodeToString(containerRegistryNode, containerRegistryNode.Content[0].Column)
			if err != nil {
				return fmt.Errorf("error while parsing containerRegistry in %s file: %s", path, err)
			}

			lines[containerRegistryNode.Line-1] = outputLine
		}

		// retag images
		if targetTag != "" {
			imagesNode := getYamlNode(globalNode, "images")
			if imagesNode != nil {
				err = updateImages(imagesNode, targetTag, lines)
				if err != nil {
					return fmt.Errorf("error while parsing images in %s file: %s", path, err)
				}
			}

			testImagesNode := getYamlNode(globalNode, "testImages")
			if testImagesNode != nil {
				err = updateImages(testImagesNode, targetTag, lines)
				if err != nil {
					return fmt.Errorf("error while parsing testImages in %s file: %s", path, err)
				}
			}
		}

		// save updated file
		if !dryRun {
			err = saveToFile(path, lines)
			if err != nil {
				return fmt.Errorf("error while saving %s file: %s", path, err)
			}
		}

		return nil
	}
}

// trimTrailingNewline trims trailing newline character
func trimTrailingNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s[:len(s)-1]
	}
	return s
}

// yamlNodeToString convers YAML node to string, with proper tabulation
func yamlNodeToString(yamlNode *yaml.Node, column int) (string, error) {
	var outputBytes []byte
	outputBytes, err := yaml.Marshal(yamlNode)
	if err != nil {
		return "", err
	}

	// add necessary tabbing
	outputLine := trimTrailingNewline(string(outputBytes))
	outputLine = strings.Repeat(" ", column-1) + outputLine

	return outputLine, nil
}

// getYamlNode finda a node with the specified key. If the next node is a map it will be returned.
func getYamlNode(parsedYaml *yaml.Node, wantedKey string) *yaml.Node {
	for key, val := range parsedYaml.Content {
		if val.Value == wantedKey {
			// "name: value" pairs are split into two values in the Content array, this should always be true
			if key+1 < len(parsedYaml.Content) {
				return parsedYaml.Content[key+1]
			}
		}
	}
	return nil
}

// updateImages looks for "version" field in each image and replaces its content with a targetTag value
func updateImages(images *yaml.Node, targetTag string, lines []string) error {
	for _, val := range images.Content {
		if val.Tag == "!!map" {
			// loop over values in singular image
			for key, imageVal := range val.Content {
				if (imageVal.Value == "version") && (key+1 < len(val.Content)) {
					// parse just the version line
					var versionLineParsed yaml.Node
					yaml.Unmarshal([]byte(lines[imageVal.Line-1]), &versionLineParsed)

					versionLineParsed.Content[0].Content[1].Value = targetTag
					outputLines, err := yamlNodeToString(&versionLineParsed, val.Content[0].Column)
					if err != nil {
						return err
					}

					lines[imageVal.Line-1] = outputLines
				}
			}
		}
	}
	return nil
}

// saveToFile saves array of lines to a file
func saveToFile(path string, lines []string) error {
	outputData := strings.Join(lines, "\n")
	outputData += "\n"

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, []byte(outputData), info.Mode())
	if err != nil {
		return err
	}
	return nil
}
