package serviceaccount

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"google.golang.org/api/iam/v1"
)

const saResourcePrefix = "projects/-/serviceAccounts/"

//Client provides data and methods for serviceaccount package.
type Client struct {
	iamservice IAM
	prefix     string
}

type ServiceAccount struct {
	Name  string                 `yaml:"name"`
	Roles []string               `yaml:"roles,omitempty"`
	Key   *iam.ServiceAccountKey `yaml:"key,omitempty"`
}

// IAM is a mockable interface for GCP IAM API.
type IAM interface {
	//TODO: Swap arguments order to match iam service method arguments order.
	CreateSA(request *iam.CreateServiceAccountRequest, projectname string) (*iam.ServiceAccount, error)
	CreateSAKey(sa string, request *iam.CreateServiceAccountKeyRequest) (*iam.ServiceAccountKey, error)
	DeleteSA(sa string) (*iam.Empty, error)
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
		log.Printf("Created serviceaccount: %s", name)
		return sa, nil
	}
}

// safqdn should be serviceaccount mail. Pass here iam.ServiceAccount.Email returned by Client.CreateSA().
func (client *Client) CreateSAKey(safqdn string) (*iam.ServiceAccountKey, error) {
	//var gkey []byte
	//TODO: creating resource string should be package global function treated as helper function. It can be used in installer package in Cleaner.CleanALL method.
	resource := fmt.Sprintf("%s%s", saResourcePrefix, safqdn)
	request := &iam.CreateServiceAccountKeyRequest{}
	key, err := client.iamservice.CreateSAKey(resource, request)
	if err != nil {
		return nil, fmt.Errorf("When creating key for serviceaccount %s, got error: %w", safqdn, err)
	}
	log.Printf("Created key for serviceaccount: %s", safqdn)
	return key, nil
}

func (client *Client) Delete(name string) (*iam.Empty, error) {
	return client.iamservice.DeleteSA(name)
}
