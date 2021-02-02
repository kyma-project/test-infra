package main

import (
	"fmt"
	"github.com/Masterminds/sprig"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
)

type Values struct {
	ContainerPort      string `yaml:"containerPort"`
	Image              string `yaml:"image"`
	Namespace          string `yaml:"namespace"`
	Name               string `yaml:"name"`
	KymaHostname       string `yaml:"kymaHostname"`
	CmpAppName         string `yaml:"cmpAppName"`
	CmpSlackAppName    string `yaml:"cmpSlackAppName"`
	WebhookSecretValue string `yaml:"webhookSecretValue"`
	RepositoryUrl      string `yaml:"repositoryUrl"`
	FunctionBaseDir    string `yaml:"functionBaseDir"`
}

var (
	valuesFilePath   = flag.String("valuesPath", "../githubConnector/values.yaml", "Path of the values file")
	templatesDirPath = flag.String("templatesPath", "../githubConnector/templates", "Path to directory with templates")
	webhookSecret    = flag.String("webhookSecret", "", "Github webhook ")
	templates        = flag.StringSlice("templates", []string{}, "Comma separated list of template filenames to generate.")
	values           Values
)

func generateFile(templateFile os.FileInfo, values Values, templatesDirPath *string) {
	t, err := template.New(templateFile.Name()).Funcs(sprig.TxtFuncMap()).ParseFiles(path.Join(*templatesDirPath, templateFile.Name()))
	if err != nil {
		log.Fatalf("failed parse template file %s", templateFile.Name())
	}
	err = t.Execute(os.Stdout, values)
	if err != nil {
		log.Fatalf("failed execute template for template file %s", templateFile.Name())
	}
	fmt.Printf("\n---\n")
}
func main() {
	flag.Parse()

	templateFiles, err := ioutil.ReadDir(*templatesDirPath)
	if err != nil {
		log.Fatalf("cannot read template files")
	}
	valuesFile, err := ioutil.ReadFile(*valuesFilePath)
	if err != nil {
		log.Fatalf("Cannot read data file: %s", err)
	}
	err = yaml.Unmarshal(valuesFile, &values)
	if err != nil {
		log.Fatalf("Cannot parse data file yaml: %s\n", err)
	}
	if *webhookSecret != "" {
		values.WebhookSecretValue = *webhookSecret
	}
	if len(*templates) != 0 {
		for _, t := range *templates {
			templateFileInfo, err := os.Stat(path.Join(*templatesDirPath, t))
			if err != nil {

			}
			generateFile(templateFileInfo, values, templatesDirPath)
		}
	} else {
		for _, templateFile := range templateFiles {
			if !templateFile.IsDir() && strings.HasSuffix(templateFile.Name(), ".yaml") {
				generateFile(templateFile, values, templatesDirPath)
				//time.Sleep(20 * time.Second)
			}
		}
	}
}
