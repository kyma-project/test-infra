package pkg

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	k8sowners "k8s.io/test-infra/prow/repoowners"
)

type Owners struct {
	Approvers []string `json:"approvers,omitempty"`
	Reviewers []string `json:"reviewers,omitempty"`
}

func GetOwnersAliases(repoBase string) (k8sowners.RepoAliases, error) {

	path := filepath.Join(repoBase, "OWNERS_ALIASES")
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return k8sowners.ParseAliasesConfig(b)

}

func ParseOwnersFile(ownersPath string) (Owners, error) {
	b, err := ioutil.ReadFile(ownersPath)
	if err != nil {
		return Owners{}, err
	}
	ownersFile := Owners{}
	err = yaml.Unmarshal(b, &ownersFile)
	if err != nil {
		return Owners{}, err
	}
	return ownersFile, err
}
