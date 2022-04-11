package common

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseNotationFile(filePath string) (string, string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", "", err
	}
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return "", "", err
		}
		rxp, _ := regexp.Compile(`^#\s+(?P<PATH>[\/\w\-\.]+):(?P<KEY>(?:\.(?:\w+)(?:\[\d+\])?)+)`)
		match := rxp.FindStringSubmatch(line)
		if len(match) > 0 {
			res := make(map[string]string)
			for i, name := range rxp.SubexpNames() {
				if i != 0 && name != "" {
					res[name] = match[i]
				}
			}
			return res["PATH"], res["KEY"], nil
		}
	}

	err = f.Close()
	if err != nil {
		return "", "", err
	}

	return "", "", fmt.Errorf("no yaml file/key notation found")
}

func getYamlByReference(parsedYaml *yaml.Node, nodePath string) (*yaml.Node, error) {
	var err error
	keyList := strings.Split(nodePath, ".")[1:]
	yamlNode := parsedYaml
	arrayRxp, _ := regexp.Compile(`^(?P<KEY1>[\w\d]+)\[(?P<KEY2>\d+)\]`)
	for _, k := range keyList {
		match := arrayRxp.FindStringSubmatch(k)
		if len(match) > 0 {
			res := make(map[string]string)
			for i, name := range arrayRxp.SubexpNames() {
				if i != 0 && name != "" {
					res[name] = match[i]
				}
			}
			index, err := strconv.Atoi(res["KEY2"])
			if err == nil {
				yamlNode, err = getYamlNodeInMap(yamlNode, res["KEY1"])
				if err != nil {
					return &yaml.Node{}, err
				}
				yamlNode = yamlNode.Content[index]
				continue
			}
		} else {
			yamlNode, err = getYamlNodeInMap(yamlNode, k)
			if err != nil {
				return &yaml.Node{}, err
			}
		}
	}
	return yamlNode, err
}

func getYamlNodeInMap(parsedYaml *yaml.Node, wantedKey string) (*yaml.Node, error) {
	for key, val := range parsedYaml.Content {
		if val.Value == wantedKey {
			return parsedYaml.Content[key+1], nil
		}
	}
	return &yaml.Node{}, fmt.Errorf("key %s not found", wantedKey)
}

func parseYamlFile(filePath string) (*yaml.Node, error) {
	var err error
	data, _ := os.Open(filePath)
	defer data.Close()
	var parsedFile yaml.Node
	decoder := yaml.NewDecoder(data)

	err = decoder.Decode(&parsedFile)
	if err != nil {
		return &yaml.Node{}, fmt.Errorf("error while unmarshalling %s file: %s", filePath, err)
	}

	_, err = data.Seek(0, 0)
	if err != nil {
		return &yaml.Node{}, err
	}
	decoder = yaml.NewDecoder(data)
	var parsedImagesFile interface{}

	err = decoder.Decode(&parsedImagesFile)
	if err != nil {
		return &yaml.Node{}, fmt.Errorf("error while decoding %s file: %s", filePath, err)
	}

	return &parsedFile, nil
}

func UpdateYamlFile(filePath string, yamlKey string, value string) error {
	parsedFile, err := parseYamlFile(filePath)
	if err != nil {
		return err
	}

	yamlNode, err := getYamlByReference(parsedFile.Content[0], yamlKey)
	if err != nil {
		return err
	}

	yamlNode.SetString(value)
	fileToWrite, _ := os.OpenFile(filePath, os.O_WRONLY, 0644)
	encoder := yaml.NewEncoder(fileToWrite)
	err = encoder.Encode(parsedFile.Content[0])
	if err != nil {
		return err
	}
	defer encoder.Close()
	return nil
}
