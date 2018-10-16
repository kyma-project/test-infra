package main

import (
	"flag"
	"fmt"
	"os"
	"text/template"
)

func main() {
	templateName := flag.String("template", "", "path to template file")
	outputFile := flag.String("out", "", "path to output plugins.yaml")
	orgUser := flag.String("orgUser", "", "github organization or user name where kyma was forked")

	flag.Parse()
	if templateName == nil || *templateName == "" {
		fmt.Println("TemplateName cannot be empty")
		flag.Usage()
		os.Exit(1)
	}
	if outputFile == nil || *outputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	if orgUser == nil || *orgUser == "" {
		flag.Usage()
		os.Exit(1)
	}

	tpl, err := template.ParseFiles(*templateName)
	if err != nil {
		panic(err)
	}

	fOut, err := os.Create(*outputFile)

	if err != nil {
		panic(err)
	}

	defer fOut.Close()

	err = tpl.Execute(fOut, PluginsConfigInput{OrganizationOrUser: *orgUser})
	if err != nil {
		panic(err)
	}

}

type PluginsConfigInput struct {
	OrganizationOrUser string
}
