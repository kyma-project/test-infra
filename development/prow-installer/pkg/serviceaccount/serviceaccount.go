package serviceaccount

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"google.golang.org/api/iam/v1"
)

//Client provides data and methods for serviceaccount package.
type Client struct {
	iamservice IAM
	prefix     string
}

// IAM is a mockable interface for GCP IAM API.
type IAM interface {
	CreateSA(request *iam.CreateServiceAccountRequest, projectname string) (*iam.ServiceAccount, error)
}

//
func NewClient(prefix string, iamservice IAM) *Client {
	return &Client{
		iamservice: iamservice,
		prefix:     prefix,
	}
}

// Creates GKE Service Account. SA name is trimed to 30 characters per GCP limits.
func (client *Client) CreateSA(name string, project string) (*iam.ServiceAccount, error) {
	if client.prefix != "" {
		name = fmt.Sprintf("%s-%s", client.prefix, name)
	}
	name = fmt.Sprintf("%.30s", name)
	project = fmt.Sprintf("projects/%s", project)
	request := &iam.CreateServiceAccountRequest{
		AccountId: name,
	}
	sa, err := client.iamservice.CreateSA(request, project)
	if err != nil {
		log.Printf("When creating %s serviceaccount in %s project got error: %w.", name, project, err)
		return nil, err
	} else {
		log.Printf("Created service account:\n %s", name)
		return sa, nil
	}
}
