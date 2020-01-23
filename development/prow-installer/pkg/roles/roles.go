package roles

import (
	"fmt"
	"regexp"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
)

// projects management object.
type client struct {
	crmservice *crmService
	policies   map[string]*cloudresourcemanager.Policy
	//projectsrequirements *[]projectRequirements
}

type CRM interface {
	GetPolicy(projectname string, getiampolicyrequest *cloudresourcemanager.GetIamPolicyRequest) (*cloudresourcemanager.Policy, error)
	SetPolicy(projectname string, seriampolicyrequest *cloudresourcemanager.SetIamPolicyRequest) (*cloudresourcemanager.Policy, error)
}

/*
// projectRequirements holds data against which project is validated.
type projectRequirements struct {
	projectname             string            `yaml:"name"`
	requiredBindings []requiredBinding `yaml:"requiredbindings"` // bindings which must exist in project policy. These are checked before policy binding is generated.
}

// Binding which must exist in project policy.
type requiredBinding struct {
	role    string   `yaml:"role"`
	members []string `yaml:"members"`
}*/

func NewClient(crmservice *crmService) (*client, error) {
	//func NewClient(crmservice *crmService, requirementsfilepath string) (*client, error) {
	//if requirementsfilepath == ""{
	//	requirementsfilepath = "./config/mandatory-requirements.yaml"
	//}
	//projectsrequirements, err := loadRequirements(requirementsfilepath)
	//if err != nil {
	//	return nil, err
	//}
	return &client{
		crmservice: crmservice,
		policies:   make(map[string]*cloudresourcemanager.Policy),
		//projectsrequirements: projectsrequirements,
	}, nil
}

func (client *client) GetPolicy(projectname string) error {
	iampolicyrequest := &cloudresourcemanager.GetIamPolicyRequest{}
	policy, err := client.crmservice.GetPolicy(projectname, iampolicyrequest)
	if err != nil && !googleapi.IsNotModified(err) {
		return fmt.Errorf("When downloading policy for %s project got error: [%v].", projectname, err)
	} else if googleapi.IsNotModified(err) {
		return fmt.Errorf("Policy for project %s was not modified.", projectname)
	}
	client.policies[projectname] = policy
	return nil
}

// Add account to the role binding members list. If role binding doesn't exist, call project.addBinding first to create it.
func (client *client) addToRole(accountfqdn string, role string, projectname string, condition *cloudresourcemanager.Expr) error {
	for _, binding := range client.policies[projectname].Bindings {
		if binding.Role == role && binding.Condition == condition {
			binding.Members = append(binding.Members, accountfqdn)
			return nil
		}
	}
	return fmt.Errorf("Binding for role %s not found in %s project policy.", role, projectname)
}

func (client *client) AddSAtoRole(saname string, role string, projectname string, condition *cloudresourcemanager.Expr) error {
	safqdn := client.makeSafqdn(saname, projectname)
	rolefullname := client.makeRoleFullname(role)
	if client.addToRole(safqdn, rolefullname, projectname, condition) != nil {
		client.addRole(safqdn, rolefullname, projectname, condition)
	}
	return nil
}

func (client *client) makeSafqdn(saname string, projectname string) string {
	//TODO: Add checks if provided saname is safqdn, if yes if the projectname match.
	return fmt.Sprintf("serviceAccount:%s@%s.iam.gserviceaccount.com", saname, projectname)
}

func (client *client) makeRoleFullname(role string) string {
	if prefixed, _ := regexp.MatchString("^(roles|organizations)/.*", role); !prefixed {
		return fmt.Sprintf("roles/%s", role)
	}
	return role
}

func (client *client) addRole(safqdn string, rolefullname string, projectname string, condition *cloudresourcemanager.Expr) {
	client.policies[projectname].Bindings = append(client.policies[projectname].Bindings, &cloudresourcemanager.Binding{
		Role:      rolefullname,
		Members:   []string{safqdn},
		Condition: condition,
	})
}

// Calls project.generatePolicyBindings and apply new policy on GKE project.
func (client *client) SetPolicy(projectname string) error {
	iampolicyrequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: client.policies[projectname],
	}
	_, err := client.crmservice.SetPolicy(projectname, iampolicyrequest)
	if err != nil && !googleapi.IsNotModified(err) {
		return fmt.Errorf("When setting new policy for %s project got error: [%v].", projectname, err)
	} else if googleapi.IsNotModified(err) {
		return fmt.Errorf("Policy for project %s was not modified.", projectname)
	}
	return nil
}

/*
// Wrapper function for policy vaildation steps. Loads project requirements file and execute policy validation functions.
func loadRequirements(requirementsfilepath string) (*[]projectRequirements, error) {
	var projectsrequirements *[]projectRequirements
	requirements, err := ioutil.ReadFile(requirementsfilepath)
	if err != nil {
		return nil, fmt.Errorf("When reading requirements file %s got error: [%v]", requirementsfilepath, err)
	}
	err = yaml.Unmarshal(requirements, projectsrequirements)
	if err != nil {
		return nil, fmt.Errorf("When unmarshalling yaml file %s got error: [%v]", requirementsfilepath, err)
	}
	return projectsrequirements, nil
}

// Validating if all bindings exist for mandatory roles with mandatory members.
// It's to prevent removing mandatory permissions from google SA or removing all project owners.
func (client *client) validateBindings(projectname string) {
	var projectbindings map[string]*cloudresourcemanager.Binding
	for _, binding := range client.policies[projectname].Bindings {
		if projectbindings[binding.Role] = binding
	}
	for _, projectrequirements := range *client.projectsrequirements{
		if projectrequirements.projectname == projectname {

		}
	}
	for _, required := range client.requirements.requiredBindings {
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
*/
