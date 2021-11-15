package promote

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

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
		}

		// retag images
		if targetTag != "" {
			imagesNode := getYamlNode(globalNode, "images")
			if imagesNode != nil {
				updateImages(imagesNode, targetTag)
			}

			testImagesNode := getYamlNode(globalNode, "testImages")
			if testImagesNode != nil {
				updateImages(testImagesNode, targetTag)
			}
		}

		// save updated file
		err = saveToFile(path, &parsedFile)
		if err != nil {
			return err
		}

		return nil
	}
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
func updateImages(images *yaml.Node, targetTag string) {
	for _, val := range images.Content {
		if val.Tag == "!!map" {
			// loop over values in singular image
			for key, imageVal := range val.Content {
				if (imageVal.Value == "version") && (key+1 < len(val.Content)) {
					val.Content[key+1].Value = targetTag
				}
			}
		}
	}
}

// saveToFile saves parsed YAML structure to a file
func saveToFile(path string, parsedFile *yaml.Node) error {
	var updatedYaml bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&updatedYaml)
	yamlEncoder.SetIndent(2)

	if err := yamlEncoder.Encode(parsedFile); err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, updatedYaml.Bytes(), 0666); err != nil {
		return err
	}
	return nil
}
