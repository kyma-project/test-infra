package accessmanager

import (
	"context"
	"fmt"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	"log"
	"strings"
)

type AccessManager struct {
	credentialsFile string
	project         string
	prefix          string

	SaAccounts *saAccountsManager
	Policies   *policiesManager
}

type saAccountsManager struct {
	accessmanager *AccessManager
	iamservice    *iam.Service
}

type policiesManager struct {
	accessmanager               *AccessManager
	cloudresourcemanagerservice *cloudresourcemanager.Service
	Bindings                    map[string]*cloudresourcemanager.Binding
	Policy                      *cloudresourcemanager.Policy
}

func NewAccessManager(credentialsfile string, project string, prefix string) *AccessManager {
	accessmanager := &AccessManager{credentialsFile: credentialsfile, project: project, prefix: prefix}

	accessmanager.SaAccounts = accessmanager.newSaAccountsService()
	accessmanager.Policies = accessmanager.newPoliciesService()

	return accessmanager
}

func (accessmanager *AccessManager) newSaAccountsService() *saAccountsManager {
	saaccountsmanager := &saAccountsManager{accessmanager: accessmanager}
	ctx := context.Background()
	iamservice, err := iam.NewService(ctx, option.WithCredentialsFile(saaccountsmanager.accessmanager.credentialsFile))
	if err != nil {
		log.Fatalf("Error %v when creating new IAM service.", err)
	} else {
		saaccountsmanager.iamservice = iamservice
	}
	return saaccountsmanager
}

func (accessmanager *AccessManager) newPoliciesService() *policiesManager {
	policiesmanager := &policiesManager{accessmanager: accessmanager}
	ctx := context.Background()
	cloudresourcemanagerService, err := cloudresourcemanager.NewService(ctx, option.WithCredentialsFile(policiesmanager.accessmanager.credentialsFile))
	if err != nil {
		log.Fatalf("Error %v when creating new cloudresourcemanager service.", err)
	} else {
		policiesmanager.cloudresourcemanagerservice = cloudresourcemanagerService
	}
	return policiesmanager
}

func (saaccountsmanager *saAccountsManager) CreateSAAccount(accounts) *iam.ServiceAccount {
	ctx := context.Background()
	if saaccountsmanager.accessmanager.prefix != "" {
		name = fmt.Sprintf("%s-%s", saaccountsmanager.accessmanager.prefix, name)
	}
	name = fmt.Sprintf("%.30s", name)
	projectName := fmt.Sprintf("projects/%s", saaccountsmanager.accessmanager.project)
	createsaaccountrequest := iam.CreateServiceAccountRequest{
		AccountId: name,
	}
	sa, err := saaccountsmanager.iamservice.Projects.ServiceAccounts.Create(projectName, &createsaaccountrequest).Context(ctx).Do()
	if err != nil && !googleapi.IsNotModified(err) {
		log.Printf("Error %v when creating new service account.", err)
	} else {
		log.Printf("Created service account:\n %s", sa.Name)
	}
	return sa
}

func (policiesmanager *policiesManager) GetProjectPolicy() {
	policiesmanager.Bindings = make(map[string]*cloudresourcemanager.Binding)
	iampolicyrequest := new(cloudresourcemanager.GetIamPolicyRequest)
	projectpolicy, err := policiesmanager.cloudresourcemanagerservice.Projects.GetIamPolicy(policiesmanager.accessmanager.project, iampolicyrequest).Do()
	if err != nil && !googleapi.IsNotModified(err) {
		log.Fatalf("Error %v when getting %s project policy.", err, policiesmanager.accessmanager.project)
	}
	policiesmanager.Policy = projectpolicy
	for _, binding := range policiesmanager.Policy.Bindings {
		rolename := strings.TrimPrefix(binding.Role, "roles/")
		policiesmanager.Bindings[rolename] = binding
	}
}

func SetSARoles(projectPolicy *cloudresourcemanager.Policy) (projectPolicy *cloudresourcemanager.Policy) {

}
