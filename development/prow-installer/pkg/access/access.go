package gcpaccessmanager

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

// Access management object.
//type AccessManager struct {
//	credentialsFile string
//	iam             *iamManager
//	projects        *projectsManager
//	Prefix          string
//	ctx             context.Context
//}

// projects management object.
type projectsManager struct {
	accessmanager               *AccessManager
	cloudresourcemanagerservice *cloudresourcemanager.Service
	projects                    map[string]*Project
	requirementsFile            string // Path to the file with projects requirements to validate. Defaults to ./config/mandatory-requirements.yaml
}

// GKE project object.
type Project struct {
	projectsManager *projectsManager
	gkeProject      *cloudresourcemanager.Project
	bindings        map[string]*cloudresourcemanager.Binding // bindings map for adding role members and new bindings to avoid multiple searches in policy bindings list.
	policy          *cloudresourcemanager.Policy             // policy generated after placing changes.
}

// projectRequirements holds data against which project is validated.
type projectRequirements struct {
	name             string            `yaml:"name"`
	requiredBindings []requiredBinding `yaml:"requiredbindings"` // bindings which must exist in project policy. These are checked before policy binding is generated.
}

// Binding which must exist in project policy.
type requiredBinding struct {
	role    string   `yaml:"role"`
	members []string `yaml:"members"`
}

// NewAccessManager returns AccessManager instance, wrapping iamManager and projectsManager for managing project permissions.
// Expects path to the gcp json credentials file. Same as value of GOOGLE_APPLICATION_CREDENTIALS.
//func NewAccessManager(credentialsfile string, prefix string) *AccessManager {
//	ctx = context.Background()
//	accessmanager := &AccessManager{credentialsFile: credentialsfile, Prefix: prefix}
//	log.Printf("AccessManager created.")
//	accessmanager.iam = accessmanager.newIAMManager()
//	accessmanager.projects = accessmanager.newProjectsManager()
//	return accessmanager
//}

// newProjectsManager establish cloudresourcemanager service client and returns ProjectManager instance with it.
func (accessManager *AccessManager) newProjectsManager() *projectsManager {
	projectsmanager := &projectsManager{accessmanager: accessManager, requirementsFile: "./config/mandatory-requirements.yaml"}
	projectsmanager.projects = make(map[string]*Project)
	cloudresourcemanagerService, err := cloudresourcemanager.NewService(accessManager.ctx, option.WithCredentialsFile(projectsmanager.accessmanager.credentialsFile))
	if err != nil {
		log.Fatalf("Error %v when creating new cloudresourcemanager service.", err)
	} else {
		projectsmanager.cloudresourcemanagerservice = cloudresourcemanagerService
		log.Printf("CloudresourcemanagerService client created.")
	}
	return projectsmanager
}

func (projectsManager *projectsManager) getProject(projectname string) (project *Project) {
	gkeproject, err := projectsManager.cloudresourcemanagerservice.Projects.Get(projectname).Context(projectsManager.accessmanager.ctx).Do()
	if err != nil {
		log.Fatalf("Error %v when getting %s project details.", err, projectname)
	} else {
		projectsManager.projects[projectname] = &Project{projectsManager: projectsManager}
		projectsManager.projects[projectname].gkeProject = gkeproject
		log.Printf("Downloaded project: %s ", gkeproject.Name)
	}
	return projectsManager.projects[projectname]
}

func (project *Project) getPolicy() {
	project.bindings = make(map[string]*cloudresourcemanager.Binding)
	iampolicyrequest := new(cloudresourcemanager.GetIamPolicyRequest)
	projectpolicy, err := project.projectsManager.cloudresourcemanagerservice.Projects.GetIamPolicy(project.gkeProject.Name, iampolicyrequest).Context(project.projectsManager.accessmanager.ctx).Do()
	if err != nil && !googleapi.IsNotModified(err) {
		log.Fatalf("Error %v when getting %s project policy.", err, project.gkeProject.Name)
	}
	project.policy = projectpolicy
	log.Printf("Downloaded policy for project: %s", project.gkeProject.Name)
	for _, binding := range project.policy.Bindings {
		rolename := strings.TrimPrefix(binding.Role, "roles/")
		if _, ok := project.bindings[rolename]; ok {
			log.Fatalf("Binding for role %s already exist. Check if there are multiple bindings with conditions for this role.", rolename)
		} else {
			project.bindings[rolename] = binding
		}
	}
}

// Add account to the role binding members list. If role binding doesn't exist, call project.addBinding first to create it.
func (project *Project) assignRole(accountfqdn string, role string) {
	if _, present := project.bindings[role]; present {
		project.bindings[role].Members = append(project.bindings[role].Members, accountfqdn)
		log.Printf("Added %s to %s role.", accountfqdn, role)
	} else {
		log.Printf("Missing binding for role: %s", role)
		project.addBinding(role)
		project.bindings[role].Members = append(project.bindings[role].Members, accountfqdn)
		log.Printf("Added %s to %s role.", accountfqdn, role)
	}
}

func (serviceAccount *ServiceAccount) MakeSAFQDN(project *Project) (safqdn string) {
	safqdn = fmt.Sprintf("serviceAccount:%s@%s.iam.gserviceaccount.com", serviceAccount.Name, project.gkeProject.Name)
	return
}

// Function generate valid binding name, create it and add it to the bindings map.
func (project *Project) addBinding(role string) {
	var roleName string
	if prefixed, _ := regexp.MatchString("^(roles|organizations)/.*", role); !prefixed {
		log.Printf("Missing roles or organization prefix in role: %s rolename.", role)
		roleName = fmt.Sprintf("roles/%s", role)
		log.Printf("Added roles prefix")
	}
	project.bindings[role] = &cloudresourcemanager.Binding{Role: roleName}
	log.Printf("Added new binding for role: %s", role)
}

// Function generate bindings list from bindings map and assign it to the policy.
func (project *Project) generatePolicyBindings() {
	var bindings []*cloudresourcemanager.Binding
	// Calling validation function to make sure all mandatory policy items are present.
	project.validatePolicy(project.projectsManager.requirementsFile)
	for _, value := range project.bindings {
		bindings = append(bindings, value)
		log.Printf("Added binding for role: %s", value.Role)
	}
	project.policy.Bindings = bindings
	log.Printf("Generated policy bindings for project: %s", project.gkeProject.Name)
}

// Calls project.generatePolicyBindings and apply new policy on GKE project.
func (project *Project) setIamPolicy() {
	project.generatePolicyBindings() // Calling here so policy bindings list is up to date and validated before applying on GKE.
	setiampolicyrequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: project.policy,
	}
	_, err := project.projectsManager.cloudresourcemanagerservice.Projects.SetIamPolicy(project.gkeProject.Name, setiampolicyrequest).Context(project.projectsManager.accessmanager.ctx).Do()
	if err != nil {
		log.Fatalf("Error %v when updating %s project policy.", err, project.gkeProject.Name)
	}
	log.Printf("Applied new policy on GKE project: %s", project.gkeProject.Name)
}

// Wrapper function for policy vaildation steps. Loads project requirements file and execute policy validation functions.
func (project *Project) validatePolicy(requirementsFilePath string) {
	var projectRequirements projectRequirements
	requirements, err := ioutil.ReadFile(requirementsFilePath)
	if err != nil {
		log.Fatalf("Error %v when reading file %s", err, requirementsFilePath)
	}
	err = yaml.Unmarshal(requirements, projectRequirements)
	if err != nil {
		log.Fatalf("Error %v when unmarshalling yaml file.", err)
	}
	log.Printf("Loaded project requirements.")
	// Call validation functions here.
	projectRequirements.validateBindings(project)
}

// Validating if all bindings exist for mandatory roles with mandatory members.
// It's to prevent removing mandatory permissions from google SA or removing all user or sa accounts.
func (projectRequirements *projectRequirements) validateBindings(project *Project) {
	for _, required := range projectRequirements.requiredBindings {
		if _, present := project.bindings[required.role]; !present {
			log.Printf("Mssing mandatory role: %s", required.role)
			project.addBinding(required.role)
			log.Printf("Added mandatory role: %s", required.role)
		}
		for _, requiredMember := range required.members {
			size := len(project.bindings[required.role].Members)
			for i, member := range project.bindings[required.role].Members {
				if requiredMember == member {
					break
				} else if (i + 1) == size {
					log.Printf("Missing required member: %s", requiredMember)
					project.assignRole(member, required.role)
					log.Printf("Added missing member: %s to the role: %s", requiredMember, required.role)
				}
			}

		}
	}
}

// Downloads GKE project and retrieve it's iam policy.
func (accessManager *AccessManager) GetGKEProject(projectName string) {
	project := accessManager.projects.getProject(projectName)
	project.getPolicy()
}

// Wrapper function implementing logic for creating GKE service account.
// It will create SA and assign SA to the specified roles.
func (accessManager *AccessManager) AddServiceAccount(project *Project, sa ServiceAccount) {
	accessManager.iam.createSAAccount(sa.Name, project.gkeProject.Name)
	safqdn := sa.MakeSAFQDN(project)
	for _, role := range sa.Roles {
		project.assignRole(safqdn, role)
	}
}

// Apply iam policy on a GKE project.
func (accessManager *AccessManager) CommitIAMPolicy(projectName string) {
	accessManager.projects.projects[projectName].setIamPolicy()
}
