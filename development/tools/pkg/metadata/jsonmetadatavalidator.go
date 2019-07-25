package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const (
	ElID          = "$id"
	ElSchema      = "$schema"
	ElProperties  = "properties"
	ElType        = "type"
	ElDescription = "description"
	ElDefault     = "default"
	ElExamples    = "examples"
)

var (
	ErrNoID             = errors.New("no '$id' element found")
	ErrNoSchema         = errors.New("no '$schema' element found")
	ErrNoDescription    = errors.New("no 'description' element found")
	ErrNoRootProperties = errors.New("no 'properties' found in the root")
	ErrNoType           = errors.New("no 'type' element found")
	ErrNoDefault        = errors.New("no 'default' element found")
	ErrNoExamples       = errors.New("no 'examples' element found")
)

func main() {

	flag.Usage = func() {
		fmt.Println("jsnschema-custom-validator <file-1.schema.json> <file-2.schema.json> ... <file-N.schema.json>")
		flag.PrintDefaults()
	}
	flag.Parse()


	fileList := os.Args[1:]
	if len(fileList) < 1 {
		fmt.Println("No input files")
		os.Exit(1)
	}

	failed := false
	for _, f := range fileList {
		fmt.Println("Validating file: " + f)
		err, errMap := validateFile(f)
		if err != nil {
			failed = true
			fmt.Printf("Failed to process file '%s': %s\r\n", f, err.Error())
			continue
		}
		if len(errMap) != 0 {
			failed = true
			for field, errs := range errMap {
				fmt.Println("Field: ", field)
				for _, e := range errs {
					fmt.Println("\t-", e.Error())
				}
			}
		}
	}

	if failed {
		os.Exit(1)
	}
}

func validateFile(fileName string) (error, map[string][]error) {
	hFile, err := os.Open(fileName)
	if err != nil {
		return errors.Wrap(err, "unable to open file "+fileName), nil
	}

	var decoded map[string]interface{}
	err = json.NewDecoder(hFile).Decode(&decoded)
	if err != nil {
		return errors.Wrap(err, "unable to decode file "+fileName), nil
	}

	err = validateRoot(decoded)
	if err != nil {
		return errors.Wrap(err, "unable to validate root element"), nil
	}

	errorsMap := make(map[string][]error)
	if props, ok := decoded[ElProperties].(map[string]interface{}); ok {
		validateProperties(props, "root", errorsMap)
	} else {
		return errors.New("unable to decode root properties"), nil
	}

	if len(errorsMap) == 0 {
		return nil, nil
	}

	return nil, errorsMap
}

func validateRoot(m map[string]interface{}) error {
	if _, ok := m[ElID]; !ok {
		return ErrNoID
	}

	if _, ok := m[ElSchema]; !ok {
		return ErrNoSchema
	}

	if _, ok := m[ElProperties]; !ok {
		return ErrNoRootProperties
	}
	return nil
}

func validateProperties(m map[string]interface{},
	fullPath string,
	e map[string][]error) {
	for k, v := range m {
		if casted, ok := v.(map[string]interface{}); ok {
			if len(casted) > 0 {
				relPath := fullPath + " - " + k
				validateElement(k, relPath, casted, e)
			}
		}
	}
}

func validateElement(eName string, fullPath string, m map[string]interface{}, e map[string][]error) {
	propsFound := false
	for k, v := range m {
		if strings.ToLower(k) == ElProperties {
			if casted, ok := v.(map[string]interface{}); ok {
				validateProperties(casted, fullPath, e)
			}
			propsFound = true
		}
	}
	if !propsFound {
		if _, ok := e[fullPath]; !ok {
			e[fullPath] = []error{}
		}
		if _, ok := m[ElType]; !ok {
			e[fullPath] = append(e[fullPath], ErrNoType)
		}
		if _, ok := m[ElDescription]; !ok {
			e[fullPath] = append(e[fullPath], ErrNoDescription)
		}
		if _, ok := m[ElDefault]; !ok {
			e[fullPath] = append(e[fullPath], ErrNoDefault)
		}
	}
	return
}