package promote

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func GetWalkFunc(ResourcesDirectoryClean, targetContainerRegistry, targetTag string) filepath.WalkFunc {
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
		lines := make([]string, 0)

		yamlFile, err := os.Open(path)
		if err != nil {
			return errors.New(err.Error() + " in file " + path)
		}
		defer yamlFile.Close()

		decoder := yaml.NewDecoder(yamlFile)
		err = decoder.Decode(&parsedFile)
		if err != nil {
			return errors.New(err.Error() + " in file " + path)
		}

		// rewind and load the file into array of lines
		yamlFile.Seek(0, 0)
		scanner := bufio.NewScanner(yamlFile)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if scanner.Err() != nil {
			return scanner.Err()
		}

		globalNode := getYamlNode(parsedFile.Content[0], "global")
		if globalNode == nil {
			return nil
		}

		// promote container registry
		if targetContainerRegistry != "" {
			containerRegistryNode := getYamlNode(globalNode, "containerRegistry")
			if containerRegistryNode == nil {
				return nil
			}

			containerRegistryPathNode := getYamlNode(containerRegistryNode, "path")
			if containerRegistryPathNode == nil {
				// TODO maybe we need some verbose info here?
				return nil
			}
			containerRegistryPathNode.Value = targetContainerRegistry

			outputLine, err := getStringFromYaml(containerRegistryNode, containerRegistryNode.Content[0].Column)
			if err != nil {
				return err
			}

			lines[containerRegistryNode.Line-1] = TrimTrailingNewline(outputLine)
		}

		// retag images
		if targetTag != "" {
			imagesNode := getYamlNode(globalNode, "images")
			if imagesNode != nil {
				updateImages(imagesNode, targetTag, lines)
			}

			testImagesNode := getYamlNode(globalNode, "testImages")
			if testImagesNode != nil {
				updateImages(testImagesNode, targetTag, lines)
			}
		}

		// save updated file
		err = saveToFile(path, lines)
		if err != nil {
			return err
		}

		return nil
	}
}

func TrimTrailingNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s[:len(s)-1]
	}
	return s
}

func getStringFromYaml(yamlNode *yaml.Node, column int) (string, error) {
	var outputBytes []byte
	outputBytes, err := yaml.Marshal(yamlNode)
	if err != nil {
		return "", err
	}

	// TODO is there a better way to push it into correct place?
	outputLine := string(outputBytes)
	for i := 1; i < column; i++ {
		outputLine = " " + outputLine
	}
	return outputLine, nil
}

// getYamlNode finda a node with the specified key. If the next node is a map it will be returned.
func getYamlNode(parsedYaml *yaml.Node, wantedKey string) *yaml.Node {
	//var tmpNode *yaml.Node
	//parsedYaml.Decode(tmpNode)
	for key, val := range parsedYaml.Content {
		if val.Value == wantedKey {
			// "name: value" pairs are split into two values in the Content array
			// TODO is this check really needed? If this is false, it should've failed on the unmarshalling step anyway
			if key+1 < len(parsedYaml.Content) {
				return parsedYaml.Content[key+1]
			}
		}
	}
	return nil
}

// updateImages looks for "version" field for each image and replaces its content with a targetTag value
func updateImages(images *yaml.Node, targetTag string, lines []string) error {
	for _, val := range images.Content {
		if val.Tag == "!!map" {
			// loop over values in singular image
			for key, imageVal := range val.Content {
				if (imageVal.Value == "version") && (key+1 < len(val.Content)) {
					// Get this particual line
					var versionLineParsed yaml.Node
					yaml.Unmarshal([]byte(lines[imageVal.Line-1]), &versionLineParsed)

					versionLineParsed.Content[0].Content[1].Value = targetTag
					outputLines, err := getStringFromYaml(&versionLineParsed, val.Content[0].Column)
					if err != nil {
						return err
					}

					lines[imageVal.Line-1] = TrimTrailingNewline(outputLines)
				}
			}
		}
	}
	return nil
}

// saveToFile saves parsed YAML structure to a file
func saveToFile(path string, lines []string) error {
	// if err := yamlEncoder.Encode(parsedFile); err != nil {
	// 	return err
	// }

	var outputData string
	for _, line := range lines {
		outputData += line + "\n"
	}
	if err := ioutil.WriteFile(path, []byte(outputData), 0666); err != nil {
		return err
	}
	return nil
}
