package serviceaccount

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
)

//Client provides data and methods for serviceaccount package.
type Client struct {
	iamservice IAM
	prefix     string
}

// IAM is a mockable interface for GCP IAM API.
type IAM interface {
	CreateSA(saname string, projectname string) (*iam.ServiceAccount, error)
}

// GCP serviceaccount options.
type SAOptions struct {
	Name    string   `yaml:"name"`
	Roles   []string `yaml:"roles,omitempty"`
	Project string   `yaml:"project"`
}

//
func NewClient(prefix string, iamservice IAM) *Client {
	return &Client{
		iamservice: iamservice,
		prefix:     prefix,
	}
}

// Creates GKE Service Account. SA name is trimed to 30 characters per GCP limits.
// If AccessManager has non zero value of prefix field, created ServiceAccounts are prefixed.
func (client *Client) CreateSA(options SAOptions) (*iam.ServiceAccount, error) {
	if client.prefix != "" {
		options.Name = fmt.Sprintf("%s-%s", client.prefix, options.Name)
	}
	options.Name = fmt.Sprintf("%.30s", options.Name)
	options.Project = fmt.Sprintf("projects/%s", options.Project)
	sa, err := client.iamservice.CreateSA(options.Name, options.Project)
	if err != nil && !googleapi.IsNotModified(err) {
		log.Printf("Error: {%v} when creating %s service account in %s project.", err, options.Name, options.Project)
		return &iam.ServiceAccount{}, err
	} else {
		log.Printf("Created service account:\n %s", options.Name)
		return sa, nil
	}
}
