package installer

import (
	"log"
	"os"
	"text/template"
)

type prowConfig struct {
	clusterip      string
	gcsprowbucket  string
	gcscredentials string
	templatePath   string
}

func NewProwConfig() *prowConfig {
	return &prowConfig{}
}

func (p *prowConfig) WithClusterIP(clusterip string) *prowConfig {
	p.clusterip = clusterip
	return p
}

func (p *prowConfig) WithGCSprowBucket(gcsprowbucket string) *prowConfig {
	p.gcsprowbucket = gcsprowbucket
	return p
}

func (p *prowConfig) WithGCScredentials(gcscredentials string) *prowConfig {
	p.gcscredentials = gcscredentials
	return p
}

func (p *prowConfig) WithTemplate(templatepath string) *prowConfig {
	p.templatePath = templatepath
	return p
}
func (p *prowConfig) GenerateConfig() {
	t, err := template.ParseFiles(p.templatePath)
	if err != nil {
		log.Print(err)
		return
	}
	f, err := os.Create("/tmp/prow_config.yaml")
	defer f.Close()
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = t.Execute(f, p)
	if err != nil {
		log.Print("execute: ", err)
		return
	}
}
