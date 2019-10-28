package accessmanager

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/installer"
)

type AccessManager struct {
	credentialsFile string

	IAM      *iamManager
	Projects *projectsManager
}

type iamManager struct {
	accessmanager *AccessManager
	iamservice    *iam.Service
}

type projectsManager struct {
	accessmanager               *AccessManager
	cloudresourcemanagerservice *cloudresourcemanager.Service
	Projects                    map[string]*Project
}

type Project struct {
	projectsManager *projectsManager
	gkeProject      *cloudresourcemanager.Project
	Bindings        map[string]*cloudresourcemanager.Binding
	Policy          *cloudresourcemanager.Policy
}

func NewAccessManager(credentialsfile string) *AccessManager {
	accessmanager := &AccessManager{credentialsFile: credentialsfile}

	accessmanager.IAM = accessmanager.newIAMService()
	accessmanager.Projects = accessmanager.newProjectsService()

	return accessmanager
}

func (accessmanager *AccessManager) newIAMService() *iamManager {
	iammanager := &iamManager{accessmanager: accessmanager}
	ctx := context.Background()
	iamservice, err := iam.NewService(ctx, option.WithCredentialsFile(iammanager.accessmanager.credentialsFile))
	if err != nil {
		log.Fatalf("Error %v when creating new IAM service.", err)
	} else {
		iammanager.iamservice = iamservice
	}
	return iammanager
}

func (accessmanager *AccessManager) newProjectsService() *projectsManager {
	projectsmanager := &projectsManager{accessmanager: accessmanager}
	projectsmanager.Projects = make(map[string]*Project)
	ctx := context.Background()
	cloudresourcemanagerService, err := cloudresourcemanager.NewService(ctx, option.WithCredentialsFile(projectsmanager.accessmanager.credentialsFile))
	if err != nil {
		log.Fatalf("Error %v when creating new cloudresourcemanager service.", err)
	} else {
		projectsmanager.cloudresourcemanagerservice = cloudresourcemanagerService
	}
	return projectsmanager
}

func (iammanager *iamManager) CreateSAAccount(name string, projectname string) *iam.ServiceAccount {
	ctx := context.Background()
	name = fmt.Sprintf("%.30s", name)
	projectName := fmt.Sprintf("projects/%s", projectname)
	createsaaccountrequest := iam.CreateServiceAccountRequest{
		AccountId: name,
	}
	sa, err := iammanager.iamservice.Projects.ServiceAccounts.Create(projectName, &createsaaccountrequest).Context(ctx).Do()
	if err != nil && !googleapi.IsNotModified(err) {
		log.Printf("Error %v when creating new service account.", err)
	} else {
		log.Printf("Created service account:\n %s", sa.Name)
	}
	return sa
}

func (projectsmanager *projectsManager) getProject(projectname string) {
	ctx := context.Background()
	project, err := projectsmanager.cloudresourcemanagerservice.Projects.Get(projectname).Context(ctx).Do()
	if err != nil {
		log.Fatalf("Error %v when geting %s project details.", err, projectname)
	} else {
		projectsmanager.Projects[projectname] = &Project{projectsManager: projectsmanager}
		projectsmanager.Projects[projectname].gkeProject = project
	}
}

func (projectsmanager *projectsManager) GetProjectPolicy(projectname string) {
	if _, present := projectsmanager.Projects[projectname]; !present {
		projectsmanager.getProject(projectname)
	}
	projectsmanager.Projects[projectname].getPolicy()
}

func (project *Project) getPolicy() {
	project.Bindings = make(map[string]*cloudresourcemanager.Binding)
	iampolicyrequest := new(cloudresourcemanager.GetIamPolicyRequest)
	projectpolicy, err := project.projectsManager.cloudresourcemanagerservice.Projects.GetIamPolicy(project.gkeProject.Name, iampolicyrequest).Do()
	if err != nil && !googleapi.IsNotModified(err) {
		log.Fatalf("Error %v when getting %s project policy.", err, project.gkeProject.Name)
	}
	project.Policy = projectpolicy
	for _, binding := range project.Policy.Bindings {
		rolename := strings.TrimPrefix(binding.Role, "roles/")
		if _, ok := project.Bindings[rolename]; ok {
			log.Fatalf("Binding for role %s already exist. Check if there are multiple bindings with conditions for this role.", rolename)
		} else {
			project.Bindings[rolename] = binding
		}
	}
}

func (projectsmanager *projectsManager) AssignRoles(projectname string, accounts []installer.Account) {
	for _, account := range accounts {
		var accountfqdn string
		if account.Type == "serviceAccount" {
			accountfqdn = fmt.Sprintf("serviceAccount:%s@%s.iam.gserviceaccount.com", account.Name, projectname)
		} else if account.Type == "user" {
			accountfqdn = fmt.Sprintf("user:%s@sap.com", account.Name)

			if _, exist := projectsmanager.Projects[projectname]; exist {
				for _, role := range account.Roles {
					if _, ok := projectsmanager.Projects[projectname].Bindings[role]; ok {
						projectsmanager.Projects[projectname].Bindings[role].Members = append(projectsmanager.Projects[projectname].Bindings[role].Members, accountfqdn)
					} else {
						bindingrole := fmt.Sprintf("roles/%s", role)
						projectsmanager.Projects[projectname].Bindings[role] = &cloudresourcemanager.Binding{Role: bindingrole, Members: []string{accountfqdn}}
					}
				}
			}
		}
		//TODO: Move this to separate function and call it from here.
		var bindings []*cloudresourcemanager.Binding
		for _, value := range projectsmanager.Projects[projectname].Bindings {
			bindings = append(bindings, value)
			projectsmanager.Projects[projectname].Policy.Bindings = bindings
		}
		ctx := context.Background()
		setiampolicyrequest := &cloudresourcemanager.SetIamPolicyRequest{
			Policy: projectsmanager.Projects[projectname].Policy,
		}
		_, err := projectsmanager.cloudresourcemanagerservice.Projects.SetIamPolicy(projectname, setiampolicyrequest).Context(ctx).Do()
		if err != nil {
			log.Fatalf("Error %v when updating %s project policy.", err, projectname)
		}
	}
}
