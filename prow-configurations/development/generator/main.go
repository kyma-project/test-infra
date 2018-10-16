package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"text/template"
)

func main() {
	templateName := flag.String("template", "", "path to template file")
	outputName := flag.String("out", "", "path to output plugins.yaml")
	inputName := flag.String("input", "", "path to JSON file to parametrize plugins template")

	flag.Parse()
	if templateName == nil || *templateName == "" {
		panic("template param cannot be empty")
	}
	if outputName == nil || *outputName == "" {
		panic("out param cannot be empty")
	}

	if inputName == nil || *inputName == "" {
		panic("input param cannot be empty")
	}

	if err := generate(*inputName,*templateName,*outputName); err != nil {
		panic(err)
	}

}

func generate(inputName, templateName, outputName string) error {
	tpl, err := template.ParseFiles(templateName)
	if err != nil {
		panic(err)
	}

	fOut, err := os.Create(outputName)

	if err != nil {
		return err
	}

	defer fOut.Close()

	fIn, err := os.Open(inputName)
	if err != nil {
		return err
	}
	defer fIn.Close()
	bytes, err := ioutil.ReadAll(fIn)
	if err != nil {
		return err
	}
	var pluginsConfig PluginsConfigInput
	if err := json.Unmarshal(bytes, &pluginsConfig); err != nil {
		return err
	}

	if err = tpl.Execute(fOut, pluginsConfig); err != nil {
		return err
	}
	if err != nil {
		panic(err)
	}

	return nil
}

type PluginsConfigInput struct {
	OrganizationOrUser string
}
