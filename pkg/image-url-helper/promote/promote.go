package promote

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kyma-project/test-infra/pkg/image-url-helper/images"

	"gopkg.in/yaml.v3"
)

// ExcludesMap contains a map of excluded filenames

type ExcludesMap map[string]bool

// ExcludesMap contains a lis of excluded filenames, used for file paring
type ExcludesList struct {
	Excludes []string `yaml:"excludes"`
}

func GetWalkFunc(ResourcesDirectoryClean, targetContainerRegistry, targetTag string, dryRun bool, images, testImages images.ComponentImageMap, excludes ExcludesMap) filepath.WalkFunc {
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

		// skip excluded values.yaml files
		if isFileExcluded(ResourcesDirectoryClean, path, excludes) {
			return nil
		}

		var parsedFile yaml.Node
		var parsedImagesFile images.ValueFile
		lines := make([]string, 0)

		yamlFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error while opening %s file: %s", path, err)
		}
		defer yamlFile.Close()

		// parse the file to a generic YAML struct with most information intact
		decoder := yaml.NewDecoder(yamlFile)
		err = decoder.Decode(&parsedFile)
		if err != nil {
			return fmt.Errorf("error while unmarshalling %s file: %s", path, err)
		}

		// load the file into array of lines
		yamlFile.Seek(0, 0)
		scanner := bufio.NewScanner(yamlFile)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if scanner.Err() != nil {
			return fmt.Errorf("error while reading %s file: %s", path, scanner.Err())
		}

		// parse the file as an easy to iterate list of images
		yamlFile.Seek(0, 0)
		decoder = yaml.NewDecoder(yamlFile)
		err = decoder.Decode(&parsedImagesFile)
		if err != nil {
			return fmt.Errorf("error while decoding %s file: %s", path, err)
		}

		// generate list of used images and apprend it to the global list containing images from all values.yaml files
		images.AppendImagesToMap(parsedImagesFile, images, testImages, "")

		globalNode := getYamlNode(parsedFile.Content[0], "global")
		if globalNode == nil {
			//no "global:" key, skip the whole file
			return nil
		}

		// promote container registry, this is always true, as this flag is required
		if targetContainerRegistry != "" {
			skip, err := promoteContainerRegistry(path, globalNode, targetContainerRegistry, lines)
			// check if the error was set or if we should skip the file even with the nil error
			if skip || (err != nil) {
				return err
			}
		}

		// retag images if the --target-tag is set
		if targetTag != "" {
			err = promoteTargetTags(path, globalNode, targetTag, lines)
			if err != nil {
				return err
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

func isFileExcluded(ResourcesDirectoryClean, path string, excludes ExcludesMap) bool {
	searchedFilename := strings.ReplaceAll(path, ResourcesDirectoryClean+"/", "")
	if _, ok := excludes[searchedFilename]; ok {
		return true
	}
	return false
}

// promoteContainerRegistry promotes container registry and returns information if the file should be skipped and error message
func promoteContainerRegistry(path string, globalNode *yaml.Node, targetContainerRegistry string, lines []string) (bool, error) {
	containerRegistryNode := getYamlNode(globalNode, "containerRegistry")
	if containerRegistryNode == nil {
		// skip files without "containerRegistry:" key
		return true, nil
	}

	containerRegistryPathNode := getYamlNode(containerRegistryNode, "path")
	if containerRegistryPathNode == nil {
		// raise error if the containerRegistry key is defined, but path is not, as this key expected to exist
		return false, fmt.Errorf("error in %s file: could not find global.containerRegistry.path key", path)
	}
	sourceContainerRegistry := containerRegistryPathNode.Value
	containerRegistryPathNode.Value = targetContainerRegistry + "/" + sourceContainerRegistry

	// update the global containerRegistry path
	outputLine, err := yamlNodeToString(containerRegistryNode, containerRegistryNode.Content[0].Column)
	if err != nil {
		return false, fmt.Errorf("error while parsing containerRegistry in %s file: %s", path, err)
	}

	lines[containerRegistryNode.Line-1] = outputLine

	// if an image has containerRegistyPath field set, then update it with the same value as global containerRegistry path
	// this is used by istio and other images that require special handling
	imagesNode := getYamlNode(globalNode, "images")
	if imagesNode != nil {
		err := updateImagesContainerRegistry(imagesNode, targetContainerRegistry, lines)
		if err != nil {
			return false, fmt.Errorf("error while parsing images in %s file: %s", path, err)
		}
	}

	testImagesNode := getYamlNode(globalNode, "testImages")
	if testImagesNode != nil {
		err := updateImagesContainerRegistry(testImagesNode, targetContainerRegistry, lines)
		if err != nil {
			return false, fmt.Errorf("error while parsing testImages in %s file: %s", path, err)
		}
	}
	return false, nil
}

func promoteTargetTags(path string, globalNode *yaml.Node, targetTag string, lines []string) error {
	imagesNode := getYamlNode(globalNode, "images")
	if imagesNode != nil {
		err := updateImagesVersion(imagesNode, targetTag, lines)
		if err != nil {
			return fmt.Errorf("error while parsing images in %s file: %s", path, err)
		}
	}

	testImagesNode := getYamlNode(globalNode, "testImages")
	if testImagesNode != nil {
		err := updateImagesVersion(testImagesNode, targetTag, lines)
		if err != nil {
			return fmt.Errorf("error while parsing testImages in %s file: %s", path, err)
		}
	}
	return nil
}

// trimTrailingNewline trims trailing newline character
func trimTrailingNewline(s string) string {
	if strings.HasSuffix(s, "\n") {
		return s[:len(s)-1]
	}
	return s
}

// yamlNodeToString converts YAML node to a string, with proper tabulation
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

// getYamlNode finds a node with the specified key. If the next node is a map it will be returned.
func getYamlNode(parsedYaml *yaml.Node, wantedKey string) *yaml.Node {
	for key, val := range parsedYaml.Content {
		if val.Value == wantedKey {
			// "name: value" pairs are split into two nodes in the Content array, this should always be true
			if key+1 < len(parsedYaml.Content) {
				return parsedYaml.Content[key+1]
			}
		}
	}
	return nil
}

// updateImagesContainerRegistry looks for "containerRegistryPath" field in each image and updates its content with a targetContainerRegistryPath value in the lines
func updateImagesContainerRegistry(images *yaml.Node, targetContainerRegistryPath string, lines []string) error {
	for _, val := range images.Content {
		if val.Tag == "!!map" {
			// loop over values in singular image
			for key, imageVal := range val.Content {
				if (imageVal.Value == "containerRegistryPath") && (key+1 < len(val.Content)) {
					// parse the containerRegistryPath line separately
					var containerRegistryPathLineParsed yaml.Node
					yaml.Unmarshal([]byte(lines[imageVal.Line-1]), &containerRegistryPathLineParsed)
					sourceContainerRegistryPath := containerRegistryPathLineParsed.Content[0].Content[1].Value
					containerRegistryPathLineParsed.Content[0].Content[1].Value = targetContainerRegistryPath + "/" + sourceContainerRegistryPath

					outputLines, err := yamlNodeToString(&containerRegistryPathLineParsed, val.Content[0].Column)
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

// updateImagesVersion looks for "version" field in each image and updates its content with a targetTag value in the lines slice
func updateImagesVersion(images *yaml.Node, targetTag string, lines []string) error {
	for _, val := range images.Content {
		if val.Tag == "!!map" {
			// loop over values in singular image
			for key, imageVal := range val.Content {
				if (imageVal.Value == "version") && (key+1 < len(val.Content)) {
					// parse the version line separately
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

// saveToFile saves array of lines to an existing file, overwriting its content and preserving file permissions
func saveToFile(path string, lines []string) error {
	outputData := strings.Join(lines, "\n")
	outputData += "\n"

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, []byte(outputData), info.Mode())
	if err != nil {
		return err
	}
	return nil
}

func ParseExcludes(excludesListFilename string) (ExcludesMap, error) {
	if excludesListFilename == "" {
		return nil, nil
	}

	excludesListFile, err := os.ReadFile(excludesListFilename)
	if err != nil {
		return nil, err
	}

	var excludesList ExcludesList

	if err = yaml.Unmarshal(excludesListFile, &excludesList); err != nil {
		return nil, err
	}
	excludesMap := make(ExcludesMap)
	for _, exclude := range excludesList.Excludes {
		excludesMap[exclude] = true
	}

	return excludesMap, nil
}
