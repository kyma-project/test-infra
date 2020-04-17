package roles

import (
	"fmt"
	"log"
	"regexp"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/api/cloudresourcemanager/v1"
)

//TODO: Add handling of policy version according to the comments in Policy type, see: https://godoc.org/google.golang.org/api/cloudresourcemanager/v1#Policy
// projects management object.
type Client struct {
	crmservice CRM
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=CRM -output=automock -outpkg=automock -case=underscore

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

// New return new Client and error object. Error is not used at present. Added it for future use and to support common error handling.
func New(crmservice CRM) (*Client, error) {
	return &Client{
		crmservice: crmservice,
	}, nil
}

//TODO: Change method signature to accept condition expression string instead *cloudresourcemanager.Expr object. Align cmd main.go code to pass expression string.
//AddSAtoRole will fetch policy from GCP, assign serviceaccount to roles and send policy back to GCP.
//If role binding doesn't exist it will be added to the policy.
//Check in caller if returned error is PolicyModifiedError. If yes, GCP policy was changed by other caller in the meantime.
func (client *Client) AddSAtoRole(saname string, roles []string, projectname string, condition *cloudresourcemanager.Expr) (*cloudresourcemanager.Policy, error) {
	//Test if mandatory input values are not empty
	if saname == "" || projectname == "" || len(roles) == 0 || func(roles []string) bool {
		for _, role := range roles {
			if role == "" {
				return true
			}
		}
		return false
	}(roles) {
		return nil, fmt.Errorf("One of mandatory method arguments saname, projectname ,role can not be empty. Got values. saname: %s projectname: %s roles: %v.", saname, projectname, roles)
	}
	match, err := regexp.MatchString(`^.+@.+\.iam\.gserviceaccount\.com$`, saname)
	if err != nil {
		//TODO: How to test this return?
		return nil, fmt.Errorf("When checking if provided saname match safqdn regex got error: %w.", err)
	}
	if match {
		return nil, fmt.Errorf("saname argument can not be serviceaccount fqdn. Provide only name, without domain part. Got value: %s.", saname)
	}

	//Getting current policy from GCP
	policy, err := client.getPolicy(projectname)
	if err != nil {
		return nil, fmt.Errorf("When adding role for serviceaccount %s got error: %w.", saname, err)
	}

	//
	safqdn := client.MakeSafqdn(saname, projectname)

	//Go over roles to assign
	for _, role := range roles {
		//Make valid rolename string
		rolefullname := client.makeRoleFullname(role)
		//Add service account to role binding
		err = client.addToRole(policy, safqdn, rolefullname, projectname, condition)
		if err != nil {
			if _, ok := err.(*BindingNotFoundError); ok {
				//If role binding was not found create it and add serviceacount to it.
				client.addRole(policy, safqdn, rolefullname, projectname, condition)
			} else {
				return nil, fmt.Errorf("When adding role for serviceaccount %s got error: %w.", saname, err)
			}
		}
	}

	//Send policy back to GCP
	err = client.setPolicy(policy, projectname)
	if err != nil {
		return nil, fmt.Errorf("When adding roles for serviceaccount %s got error: %w", safqdn, err)
	}
	log.Printf("Assigned %s to roles: %v", safqdn, roles)
	return policy, nil
}

//TODO: Change method signature to accept condition expression string instead *cloudresourcemanager.Expr object. Align cmd main.go code to pass expression string.
//AddSAtoRole will fetch policy from GCP, assign serviceaccount to roles and send policy back to GCP.
//If role binding doesn't exist it will be added to the policy.
//Check in caller if returned error is PolicyModifiedError. If yes, GCP policy was changed by other caller in the meantime.
func (client *Client) RemoveSaRole(saname string, roles []string, projectname string, condition *cloudresourcemanager.Expr) (*cloudresourcemanager.Policy, error) {
	//Test if mandatory input values are not empty
	if saname == "" || projectname == "" || len(roles) == 0 || func(roles []string) bool {
		for _, role := range roles {
			if role == "" {
				return true
			}
		}
		return false
	}(roles) {
		return nil, fmt.Errorf("One of mandatory method arguments saname, projectname ,role can not be empty. Got values. saname: %s projectname: %s roles: %v.", saname, projectname, roles)
	}
	match, err := regexp.MatchString(`^.+@.+\.iam\.gserviceaccount\.com$`, saname)
	if err != nil {
		//TODO: How to test this return?
		return nil, fmt.Errorf("When checking if provided saname match safqdn regex got error: %w.", err)
	}
	if match {
		return nil, fmt.Errorf("saname argument can not be serviceaccount fqdn. Provide only name, without domain part. Got value: %s.", saname)
	}

	//Getting current policy from GCP
	policy, err := client.getPolicy(projectname)
	if err != nil {
		return nil, fmt.Errorf("When removing role for serviceaccount %s got error: %w.", saname, err)
	}

	//
	safqdn := client.MakeSafqdn(saname, projectname)

	//Go over roles to assign
	for _, role := range roles {
		//Make valid rolename string
		rolefullname := client.makeRoleFullname(role)
		//Add service account to role binding
		err = client.removeFromRole(policy, safqdn, rolefullname, projectname, condition)
		if err != nil {
			if _, ok := err.(*BindingNotFoundError); ok {
				//If role binding was not found create it and add serviceacount to it.
				continue
			} else {
				return nil, fmt.Errorf("When removing role for serviceaccount %s got error: %w.", saname, err)
			}
		}
	}

	//Send policy back to GCP
	err = client.setPolicy(policy, projectname)
	if err != nil {
		return nil, fmt.Errorf("When removing roles for serviceaccount %s got error: %w", safqdn, err)
	}
	log.Printf("Removed %s from roles: %v", safqdn, roles)
	return policy, nil
}

//getPolicy will fetch policy from GCP
func (client *Client) getPolicy(projectname string) (*cloudresourcemanager.Policy, error) {
	iampolicyrequest := &cloudresourcemanager.GetIamPolicyRequest{}
	policy, err := client.crmservice.GetPolicy(projectname, iampolicyrequest)
	if err != nil {
		return nil, fmt.Errorf("When downloading policy for %s project got error: %w.", projectname, err)
	}
	return policy, nil
}

//addToRole will search role binding and add serviceaccount to the role binding members list.
func (client *Client) addToRole(policy *cloudresourcemanager.Policy, safqdn string, rolefullname string, projectname string, condition *cloudresourcemanager.Expr) error {
	for index, binding := range policy.Bindings {
		if binding.Role == rolefullname && cmp.Equal(binding.Condition, condition) {
			policy.Bindings[index].Members = append(policy.Bindings[index].Members, safqdn)
			return nil
		}
	}
	return &BindingNotFoundError{msg: fmt.Sprintf("Binding for role %s not found in %s project policy.", rolefullname, projectname)}
}

func (client *Client) removeFromRole(policy *cloudresourcemanager.Policy, safqdn string, rolefullname string, projectname string, condition *cloudresourcemanager.Expr) error {
	for index, binding := range policy.Bindings {
		if binding.Role == rolefullname && cmp.Equal(binding.Condition, condition) {
			for i, v := range binding.Members {
				if v == safqdn {
					a := binding.Members
					a[i] = a[len(a)-1]
					a[len(a)-1] = ""
					a = a[:len(a)-1]
					policy.Bindings[index].Members = a
				}
			}
			return nil
		}
	}
	return &BindingNotFoundError{msg: fmt.Sprintf("Binding for role %s not found in %s project policy.", rolefullname, projectname)}
}

//TODO: This should be renamed to make sa resource string. It should not be exported. Revert it to client private method.
//makeSafqdn will create serviceaccount fully qualified valid name, accepted by GCP API.
func (client *Client) MakeSafqdn(saname string, projectname string) string {
	return fmt.Sprintf("serviceAccount:%s@%s.iam.gserviceaccount.com", saname, projectname)
}

//makeRoleFullname will create role name valid string, accepted by GCP API.
func (client *Client) makeRoleFullname(role string) string {
	if prefixed, _ := regexp.MatchString("^(roles|organizations)/.*", role); !prefixed {
		return fmt.Sprintf("roles/%s", role)
	}
	return role
}

//addRole will create new binding for role and add serviceaccount to members list.
func (client *Client) addRole(policy *cloudresourcemanager.Policy, safqdn string, rolefullname string, projectname string, condition *cloudresourcemanager.Expr) {
	policy.Bindings = append(policy.Bindings, &cloudresourcemanager.Binding{
		Role:      rolefullname,
		Members:   []string{safqdn},
		Condition: condition,
	})
}

//setPolicy will send policy back to GCP.
//It will check if policy was not modified and differ from the one which was downloaded and modified.
//Policy modification is detected by comparing policy resource etag.
func (client *Client) setPolicy(policy *cloudresourcemanager.Policy, projectname string) error {
	currentpolicy, err := client.getPolicy(projectname)
	if err == nil {
		if currentpolicy.Etag != policy.Etag {
			return &PolicyModifiedError{msg: fmt.Sprintf("When checking if policy was modified for %s project got: Policy was modified.", projectname)}
		}
	} else {
		return fmt.Errorf("When sending new policy, failed download current policy from GCP. Got error: %w", err)
	}
	iampolicyrequest := &cloudresourcemanager.SetIamPolicyRequest{
		Policy: policy,
	}
	policy, err = client.crmservice.SetPolicy(projectname, iampolicyrequest)
	if err != nil {
		return fmt.Errorf("When setting new policy for %s project got error: %w.", projectname, err)
	}
	return nil
}
