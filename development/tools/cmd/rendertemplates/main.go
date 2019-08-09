package main

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"
	"github.com/Masterminds/sprig"
)

var (
	configFilePath = flag.String("config", "", "Path of the config file")
)

type Config struct {
	Templates []TemplateConfig
}

type TemplateConfig struct {
	From   string
	Render []RenderConfig
}

type RenderConfig struct {
	To     string
	Values map[string]interface{}
}

func main() {
	flag.Parse()

	if *configFilePath == "" {
		log.Fatal("Provide path to config file with --config")
	}

	configFile, err := ioutil.ReadFile(*configFilePath)
	if err != nil {
		log.Fatalf("Cannot read config file: %s", err)
	}

	config := new(Config)
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		log.Fatalf("Cannot parse config yaml: %s\n", err)
	}

	for _, templateConfig := range config.Templates {
		err := renderTemplate(path.Dir(*configFilePath), templateConfig)
		if err != nil {
			log.Fatalf("Cannot render template %s: %s", templateConfig.From, err)
		}
	}
}

func renderTemplate(basePath string, templateConfig TemplateConfig) error {
	templateInstance, err := loadTemplate(basePath, templateConfig.From)
	if err != nil {
		return err
	}

	for _, render := range templateConfig.Render {
		err := renderFileFromTemplate(basePath, templateInstance, render)
		if err != nil {
			return err
		}
	}

	return nil
}

func renderFileFromTemplate(basePath string, templateInstance *template.Template, renderConfig RenderConfig) error {
	relativeDestPath := path.Join(basePath, renderConfig.To)
	destFile, err := os.Create(relativeDestPath)
	if err != nil {
		return err
	}

	return templateInstance.Execute(destFile, renderConfig.Values)
}

func loadTemplate(basePath, templatePath string) (*template.Template, error) {
	relativeTemplatePath := path.Join(basePath, templatePath)
	return template.New(path.Base(templatePath)).Funcs(sprig.TxtFuncMap()).ParseFiles(relativeTemplatePath)
}
