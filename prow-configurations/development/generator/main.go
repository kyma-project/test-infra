package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/pkg/errors"
)

func main() {
	templateName := flag.String("template", "", "path to template file")
	outputName := flag.String("out", "", "path to output plugins.yaml")
	inputName := flag.String("input", "", "path to JSON file to parametrize plugins template")

	flag.Parse()
	if templateName == nil || *templateName == "" {
		panic("'template' param cannot be empty")
	}
	if outputName == nil || *outputName == "" {
		panic("'out' param cannot be empty")
	}

	if inputName == nil || *inputName == "" {
		panic("'input' param cannot be empty")
	}

	if err := generate(*inputName, *templateName, *outputName); err != nil {
		panic(err)
	}

}

func generate(inputName, templateName, outputName string) error {
	tpl, err := template.ParseFiles(templateName)
	if err != nil {
		return errors.Wrap(err, "while parsing template files")
	}

	fOut, err := os.Create(outputName)

	if err != nil {
		return errors.Wrapf(err, "while creating output file [%s]", outputName)
	}

	defer fOut.Close()

	fIn, err := os.Open(inputName)
	if err != nil {
		return errors.Wrapf(err, "while opening input file [%s]", inputName)
	}
	defer fIn.Close()
	bytes, err := ioutil.ReadAll(fIn)
	if err != nil {
		return errors.Wrapf(err, "while reading from input file")
	}
	var config ConfigInput
	if err := json.Unmarshal(bytes, &config); err != nil {
		return errors.Wrapf(err, "while unmarshalling ConfigInput")
	}

	if err = tpl.Execute(fOut, config); err != nil {
		return errors.Wrap(err, "while executing template")
	}
	return nil
}

// ConfigInput provided configuration options for Prow templates.
type ConfigInput struct {
	OrganizationOrUser string
	Plank              Plank
}

type Plank struct {
	Bucket string
}
