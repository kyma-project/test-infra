package roles

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
)

// projects management object.
type client struct {
	crmservice CRM
	policies   map[string]*cloudresourcemanager.Policy
}

type CRM interface {
	GetPolicy(projectname string, getiampolicyrequest *cloudresourcemanager.GetIamPolicyRequest) (*cloudresourcemanager.Policy, error)
	SetPolicy(projectname string, setiampolicyrequest *cloudresourcemanager.SetIamPolicyRequest) (*cloudresourcemanager.Policy, error)
}

//Custom Errors
type PolicyModifiedError struct {
	msg string //description of error
}

type BindingNotFoundError struct {
	msg string //description of error
}

// Implementation of Error interface.
func (e *PolicyModifiedError) Error() string  { return e.msg }
func (e *BindingNotFoundError) Error() string { return e.msg }

// New return new client and error object. Error is not used at present. Added it for future use and to support common error handling.
func New(crmservice CRM) (*client, error) {
	return &client{
		crmservice: crmservice,
		policies:   make(map[string]*cloudresourcemanager.Policy),
	}, nil
}

//Check in caller if returned error is PolicyModifiedError. If yes, during this method execution GCP policy was changed by other caller. Writing policy to GKE would override changes.
func (client *client) AddSAtoRole(saname string, roles []string, projectname string, condition *cloudresourcemanager.Expr) (*cloudresourcemanager.Policy, error) {
	//Test if mandatory input values are not empty
	if saname == "" || projectname == "" || len(roles) == 0 || func(roles []string) bool {
		for _, role := range roles {
			if role == "" {
				return true
			}
		}
		return false
	}(roles) {
		return nil, fmt.Errorf("One of mandatory method arguments saname, projectname ,role can not be empty. Got values. saname: [%s] projectname: [%s] roles: [%v].", saname, projectname, roles)
	}
	match, err := regexp.MatchString(`^.+@.+\.iam\.gserviceaccount\.com$`, saname)
	if err != nil {
		return nil, fmt.Errorf("When checking if provided saname match safqdn regex got error: [%w].", err)
	}
	if match {
		return nil, fmt.Errorf("saname argument can not be serviceaccount fqdn. Provide only name, without domain part. Got value: [%s].", saname)
	}
	if _, present := client.policies[projectname]; !present {
		policy, err := client.getPolicy(projectname)
		if err != nil {
			return nil, fmt.Errorf("When adding role for serviceaccount %s got error: [%w].", saname, err)
		}
		client.policies[projectname] = policy
	}
	safqdn := client.makeSafqdn(saname, projectname)
	for _, role := range roles {
		rolefullname := client.makeRoleFullname(role)
		err := client.addToRole(safqdn, rolefullname, projectname, condition)
		if err != nil {
			if _, ok := err.(*BindingNotFoundError); ok {
				client.addRole(safqdn, rolefullname, projectname, condition)
			} else {
				client.policies[projectname] = nil
				return nil, fmt.Errorf("When adding role for serviceaccount %s got error: [%w].", saname, err)
			}
		}
	}
	policy, err = client.setPolicy(projectname)
	if err != nil {
		client.policies[projectname] = nil
		return nil, fmt.Errorf("When adding roles for serviceaccount [%s] got error: [%w]", safqdn, err)
	}
	return policy, nil
}

func (client *client) getPolicy(projectname string) (*cloudresourcemanager.Policy, error) {
	iampolicyrequest := &cloudresourcemanager.GetIamPolicyRequest{}
	policy, err := client.crmservice.GetPolicy(projectname, iampolicyrequest)
	if err != nil {
		return nil, fmt.Errorf("When downloading policy for %s project got error: [%w].", projectname, err)
	}
	return policy, nil
}

// Add account to the role binding members list. If role binding doesn't exist, call project.addBinding first to create it.
func (client *client) addToRole(safqdn string, rolefullname string, projectname string, condition *cloudresourcemanager.Expr) error {
	for _, binding := range client.policies[projectname].Bindings {
		if binding.Role == rolefullname && cmp.Equal(binding.Condition, condition) {
			binding.Members = append(binding.Members, safqdn)
			return nil
		}
	}
	return &BindingNotFoundError{msg: fmt.Sprintf("Binding for role %s not found in %s project policy.", rolefullname, projectname)}
}

func (client *client) makeSafqdn(saname string, projectname string) string {
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
func (client *client) setPolicy(projectname string) (*cloudresourcemanager.Policy, error) {
	_, err := client.getPolicy(projectname)
	if err != nil && !googleapi.IsNotModified(errors.Unwrap(err)) {
		return nil, fmt.Errorf("When checking if policy was modified for %s project got error: [%w].", projectname, err)
	}
	if err == nil {
		return nil, &PolicyModifiedError{msg: fmt.Sprintf("When checking if policy was modified for [%s] project got: Policy was modified.", projectname)}
	}
	iampolicyrequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: client.policies[projectname],
	}
	policy, err := client.crmservice.SetPolicy(projectname, iampolicyrequest)
	if err != nil {
		return nil, fmt.Errorf("When setting new policy for [%s] project got error: [%w].", projectname, err)
	}
	return policy, nil
}
