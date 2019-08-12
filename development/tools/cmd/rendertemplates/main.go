package main

import (
	"flag"
	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"
)

var (
	configFilePath = flag.String("config", "", "Path of the config file")
)

type Config struct {
	Templates []TemplateConfig
	Global    map[string]interface{}
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
		err := renderTemplate(path.Dir(*configFilePath), templateConfig, config)
		if err != nil {
			log.Fatalf("Cannot render template %s: %s", templateConfig.From, err)
		}
	}
}

func renderTemplate(basePath string, templateConfig TemplateConfig, config *Config) error {
	templateInstance, err := loadTemplate(basePath, templateConfig.From)
	if err != nil {
		return err
	}

	for _, render := range templateConfig.Render {
		err := renderFileFromTemplate(basePath, templateInstance, render, config)
		if err != nil {
			return err
		}
	}

	return nil
}

func renderFileFromTemplate(basePath string, templateInstance *template.Template, renderConfig RenderConfig, config *Config) error {
	relativeDestPath := path.Join(basePath, renderConfig.To)

	destDir := path.Dir(relativeDestPath)
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return err
	}

	destFile, err := os.Create(relativeDestPath)
	if err != nil {
		return err
	}

	values := map[string]interface{}{"Values": renderConfig.Values, "Global": config.Global}

	return templateInstance.Execute(destFile, values)
}

func loadTemplate(basePath, templatePath string) (*template.Template, error) {
	relativeTemplatePath := path.Join(basePath, templatePath)
	return template.New(path.Base(templatePath)).Funcs(sprig.TxtFuncMap()).ParseFiles(relativeTemplatePath)
}
