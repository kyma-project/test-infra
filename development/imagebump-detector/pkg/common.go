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
	defer f.Close()
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(f)
	for {
		line, err2 := r.ReadString('\n')
		if err2 == io.EOF {
			break
		} else if err != nil {
			fmt.Print(fmt.Errorf("error: %s", err2))
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
	return "", "", fmt.Errorf("No yaml file/key notation found")
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
