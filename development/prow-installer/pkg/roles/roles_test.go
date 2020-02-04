package roles

import (
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/roles/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	"net/http"
	"regexp"
	"testing"
)

//TODO: Move test values definition under each test function.
//TODO: Align test values to use testvalue type.
var testvalues = []struct {
	saname    string
	project   string
	roles     []string
	policy    *cloudresourcemanager.Policy
	condition *cloudresourcemanager.Expr
}{
	{saname: "test_sa_01", project: "test_project_01", roles: []string{"test_role_01"}, policy: &cloudresourcemanager.Policy{}, condition: nil},
	{saname: "test_sa_02", project: "test_project_02", roles: []string{"test_role_02"}, policy: &cloudresourcemanager.Policy{
		AuditConfigs: nil,
		Bindings: []*cloudresourcemanager.Binding{&cloudresourcemanager.Binding{
			Condition:       nil,
			Members:         []string{"group:prow_admins@sap.com", "serviceAccount:some_sa@test_project.iam.gserviceaccount.com"},
			Role:            "roles/owner",
			ForceSendFields: nil,
			NullFields:      nil,
		}, &cloudresourcemanager.Binding{
			Condition:       nil,
			Members:         []string{"serviceAccount:service-727270599349@gs-project-accounts.iam.gserviceaccount.com", "user:some_user@sap.com"},
			Role:            "roles/compute.admin",
			ForceSendFields: nil,
			NullFields:      nil,
		}},
		Etag:            "",
		Version:         0,
		ServerResponse:  googleapi.ServerResponse{},
		ForceSendFields: nil,
		NullFields:      nil,
	}, condition: &cloudresourcemanager.Expr{
		Description:     "",
		Expression:      "test-expression",
		Location:        "",
		Title:           "",
		ForceSendFields: nil,
		NullFields:      nil,
	}},
	{saname: "test_sa_03", project: "test_project_03", roles: []string{"organizations/test_role_03"}, policy: &cloudresourcemanager.Policy{
		AuditConfigs: nil,
		Bindings: []*cloudresourcemanager.Binding{&cloudresourcemanager.Binding{
			Condition:       nil,
			Members:         []string{"group:prow_admins@sap.com", "serviceAccount:some_sa@test_project.iam.gserviceaccount.com"},
			Role:            "roles/owner",
			ForceSendFields: nil,
			NullFields:      nil,
		}, &cloudresourcemanager.Binding{
			Condition:       nil,
			Members:         []string{"serviceAccount:service-727270599349@gs-project-accounts.iam.gserviceaccount.com", "user:some_user@sap.com"},
			Role:            "roles/compute.admin",
			ForceSendFields: nil,
			NullFields:      nil,
		}},
		Etag:            "",
		Version:         0,
		ServerResponse:  googleapi.ServerResponse{},
		ForceSendFields: nil,
		NullFields:      nil,
	}, condition: nil},
	{saname: "test_sa_04", project: "test_project_04", roles: []string{"roles/test_role_04"}, policy: &cloudresourcemanager.Policy{
		AuditConfigs: nil,
		Bindings: []*cloudresourcemanager.Binding{&cloudresourcemanager.Binding{
			Condition:       nil,
			Members:         []string{"group:prow_admins@sap.com", "serviceAccount:some_sa@test_project.iam.gserviceaccount.com"},
			Role:            "roles/owner",
			ForceSendFields: nil,
			NullFields:      nil,
		}, &cloudresourcemanager.Binding{
			Condition:       nil,
			Members:         []string{"serviceAccount:service-727270599349@gs-project-accounts.iam.gserviceaccount.com", "user:some_user@sap.com"},
			Role:            "roles/test_role_04",
			ForceSendFields: nil,
			NullFields:      nil,
		}},
		Etag:            "",
		Version:         0,
		ServerResponse:  googleapi.ServerResponse{},
		ForceSendFields: nil,
		NullFields:      nil,
	}, condition: nil},
}

var notModifiedError = &googleapi.Error{
	Code:    http.StatusNotModified,
	Message: "",
	Body:    "",
	Header:  nil,
	Errors:  nil,
}

func Test_New(t *testing.T) {
	t.Run("New() should create client object without errors.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		crmclient, err := New(mockCRM)
		if test := assert.IsTypef(t, &client{}, crmclient, "\tnot expected: New() returned client object not type of *client{}."); test {
			t.Log("\texpected: New() returned expected client object.")
		}
		if test := assert.Nilf(t, err, "\tnot expected: New() returned not nil error."); test {
			t.Log("\texpected: New() returned nil error.")
		}
		if test := assert.Equalf(t, mockCRM, crmclient.crmservice, "\tnot expected: New() returned client object with unexpected crmservice field."); test {
			t.Log("\texpected: New() returned client object with correct crmservice field.")
		}
		if test := assert.Emptyf(t, crmclient.policies, "\tnot expected: New() returned client with not empty policies field."); test {
			t.Log("\texpected: New() returned client with empty policies field")
		}
	})
}

// NOT READY
// tests: without errors, each string arg empty, add sa to existing role, add sa to non existing role, add sa to role where it's already present, add sa to multiple roles at one function call, add sa to different roles in separate calls.
func TestClient_AddSAtoRole(t *testing.T) {
	type testvalue = struct {
		saname    string
		project   string
		roles     []string
		policy    *cloudresourcemanager.Policy
		condition *cloudresourcemanager.Expr
	}
	t.Run("AddSAtoRole should fail because of missing mandatory arguments", func(t *testing.T) {
		//test with empty saname, projectname, roles slice and roles members.
		values := []testvalue{
			{saname: "", project: "test_project_01", roles: []string{"test_role_01"}, policy: nil, condition: nil},
			{saname: "test_sa_01", project: "", roles: []string{"test_role_01"}, policy: nil, condition: nil},
			{saname: "test_sa_01", project: "test_project_01", roles: nil, policy: nil, condition: nil},
			{saname: "test_sa_01", project: "test_project_01", roles: []string{""}, policy: nil, condition: nil},
		}
		for _, value := range values {
			t.Logf("\n\tTesting with values:\n\tsaname: %s\n\tnproejct: %s\n\troles: %v", value.saname, value.project, value.roles)
			mockCRM := &mocks.CRM{}
			client, _ := New(mockCRM)
			//mockCRM.On("GetPolicy", value.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, nil)
			//mockCRM.On("SetPolicy", value.project, &cloudresourcemanager.SetIamPolicyRequest{Policy: nil,}).Return(nil, nil)
			defer mockCRM.AssertExpectations(t)
			policy, err := client.AddSAtoRole(value.saname, value.roles, value.project, nil)
			//should return error
			if test := assert.EqualErrorf(t, err, fmt.Sprintf("One of mandatory method arguments saname, projectname ,role can not be empty. Got values. saname: %s projectname: %s roles: %v.", value.saname, value.project, value.roles), "\tnot expected: AddSAtoRole returned unexpected error or nil."); test {
				t.Log("\texpected: AddSAtoRole() returned expected error.")
			}
			//should return nil policy
			if test := assert.Nil(t, policy, "\tnotexpected: AddSAtoRole returned not nil policy."); test {
				t.Log("\texpected: AddSAtoRole returned nil policy.")
			}
			//should not call crmservice.GetPolicy
			if test := mockCRM.AssertNotCalled(t, "GetPolicy"); test {
				t.Log("\texpected: AddSAtoRole() did not call crmservice.GetPolicy().")
			} else {
				t.Log("\tnot expected: AddSAtoRole() did call crmservice.GetPolicy().")
			}
			//should not call crmservice.SetPolicy
			if test := mockCRM.AssertNotCalled(t, "SetPolicy"); test {
				t.Log("\texpected: AddSAtoRole() did not call crmservice.SetPolicy().")
			} else {
				t.Log("\tnot expected: AddSAtoRole() did call crmservice.SetPolicy().")
			}
			//client.policies should not contain project policy
			if test := assert.Nilf(t, client.policies[value.project], "\tnot expected: Client object holds policy for project: %s.", value.project); test {
				t.Logf("\texpected: Client object do not contain policy for project: %s.", value.project)
			}
		}
	})
	t.Run("AddSAtoRole should fail because safqdn passed as saname argument", func(t *testing.T) {
		//test with safqdn passed as saname
		var (
			saname  = "test_sa_01@test_project_01.iam.gserviceaccount.com"
			project = "test_project_01"
			roles   = []string{"test_role_01"}
		)
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		//mockCRM.On("GetPolicy", value.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, nil)
		//mockCRM.On("SetPolicy", value.project, &cloudresourcemanager.SetIamPolicyRequest{Policy: nil,}).Return(nil, nil)
		defer mockCRM.AssertExpectations(t)
		policy, err := client.AddSAtoRole(saname, roles, project, nil)
		//should return error
		if test := assert.EqualErrorf(t, err, fmt.Sprintf("saname argument can not be serviceaccount fqdn. Provide only name, without domain part. Got value: %s.", saname), "\tnot expected: AddSAtoRole returned unexpected error or nil."); test {
			t.Log("\texpected: AddSAtoRole() returned expected error.")
		}
		//should return nil policy
		if test := assert.Nil(t, policy, "\tnotexpected: AddSAtoRole returned not nil policy."); test {
			t.Log("\texpected: AddSAtoRole returned nil policy.")
		}
		//should not call crmservice.GetPolicy
		if test := mockCRM.AssertNotCalled(t, "GetPolicy"); test {
			t.Log("\texpected: AddSAtoRole() did not call crmservice.GetPolicy().")
		} else {
			t.Log("\tnot expected: AddSAtoRole() did call crmservice.GetPolicy().")
		}
		//should not call crmservice.SetPolicy
		if test := mockCRM.AssertNotCalled(t, "SetPolicy"); test {
			t.Log("\texpected: AddSAtoRole() did not call crmservice.SetPolicy().")
		} else {
			t.Log("\tnot expected: AddSAtoRole() did call crmservice.SetPolicy().")
		}
		//client.policies should not contain project policy
		if test := assert.Nilf(t, client.policies[project], "\tnot expected: Client object holds policy for project: %s.", project); test {
			t.Logf("\texpected: Client object do not contain policy for project: %s.", project)
		}
	})
	t.Run("AddSAtoRole should fail because got error when getting policy from GCP", func(t *testing.T) {
		//test with correct arguments
		value := testvalue{saname: "test_sa_01", project: "test_project_01", roles: []string{"test_role_01"}, policy: &cloudresourcemanager.Policy{
			AuditConfigs: nil,
			Bindings: []*cloudresourcemanager.Binding{&cloudresourcemanager.Binding{
				Condition:       nil,
				Members:         []string{"group:prow_admins@sap.com", "serviceAccount:some_sa@test_project.iam.gserviceaccount.com"},
				Role:            "roles/owner",
				ForceSendFields: nil,
				NullFields:      nil,
			}, &cloudresourcemanager.Binding{
				Condition:       nil,
				Members:         []string{"serviceAccount:service-727270599349@gs-project-accounts.iam.gserviceaccount.com", "user:some_user@sap.com"},
				Role:            "roles/compute.admin",
				ForceSendFields: nil,
				NullFields:      nil,
			}},
			Etag:            "",
			Version:         0,
			ServerResponse:  googleapi.ServerResponse{},
			ForceSendFields: nil,
			NullFields:      nil,
		}, condition: nil}
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		mockCRM.On("GetPolicy", value.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, errors.New("GetPolicy() error."))
		defer mockCRM.AssertExpectations(t)
		policy, err := client.AddSAtoRole(value.saname, value.roles, value.project, value.condition)
		//should return error
		if test := assert.Errorf(t, err, "\tnot expected: AddSAtoRole do not returned error."); test {
			t.Log("\texpected: AddSAtoRole() returned error.")
			if test := assert.Containsf(t, err.Error(), "When downloading policy for", "\tnot expected: AddSAtoRole() returned unexpected error message."); test {
				t.Log("\texpected: AddSAtoRole() returned expected error message.")
			}
		}
		//should return nil policy
		if test := assert.Nil(t, policy, "\tnotexpected: AddSAtoRole returned not nil policy."); test {
			t.Log("\texpected: AddSAtoRole returned nil policy.")
		}
		//should call crmservice.GetPolicy
		if test := mockCRM.AssertCalled(t, "GetPolicy", value.project, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\texpected: AddSAtoRole() did call crmservice.GetPolicy().")
		} else {
			t.Log("\tnot expected: AddSAtoRole() did not call crmservice.GetPolicy().")
		}
		//should call crmservice.GetPolicy once
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\texpected: AddSAtoRole() called crmservice.GetPolicy() expected number of times.")
		} else {
			t.Log("\tnot expected: AddSAtoRole() called crmservice.GetPolicy() unexpected number of times.")
		}
		//should not call crmservice.SetPolicy
		if test := mockCRM.AssertNotCalled(t, "SetPolicy"); test {
			t.Log("\texpected: AddSAtoRole() did not call crmservice.SetPolicy().")
		} else {
			t.Log("\tnot expected: AddSAtoRole() did call crmservice.SetPolicy().")
		}
		//client.policies should not contain project policy
		if test := assert.Nilf(t, client.policies[value.project], "\tnot expected: Client object holds policy for project: %s.", value.project); test {
			t.Logf("\texpected: Client object do not contain policy for project: %s.", value.project)
		}
	})
	/*
		// TODO: Can't test this as different return values are needed for different cals to crmservice.GetPolicy method. Need research how to do it.
		t.Run("AddSAtoRole should fail because got PolicyModifiedError when setting policy in GCP", func(t *testing.T) {
			//test with correct values
			value := testvalue{saname: "test_sa_01", project: "test_project_01", roles: []string{"test_role_01"}, policy: &cloudresourcemanager.Policy{
				AuditConfigs: nil,
				Bindings: []*cloudresourcemanager.Binding{&cloudresourcemanager.Binding{
					Condition:       nil,
					Members:         []string{"group:prow_admins@sap.com", "serviceAccount:some_sa@test_project.iam.gserviceaccount.com"},
					Role:            "roles/owner",
					ForceSendFields: nil,
					NullFields:      nil,
				}, &cloudresourcemanager.Binding{
					Condition:       nil,
					Members:         []string{"serviceAccount:service-727270599349@gs-project-accounts.iam.gserviceaccount.com", "user:some_user@sap.com"},
					Role:            "roles/compute.admin",
					ForceSendFields: nil,
					NullFields:      nil,
				}},
				Etag:            "",
				Version:         0,
				ServerResponse:  googleapi.ServerResponse{},
				ForceSendFields: nil,
				NullFields:      nil,
			}, condition: nil}
			mockCRM := &mocks.CRM{}
			client, _ := New(mockCRM)
			mockCRM.On("GetPolicy", value.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, &PolicyModifiedError{msg:"test PolicyModifiedError"})
			//mockCRM.On("SetPolicy", value.project, &cloudresourcemanager.SetIamPolicyRequest{
			//	Policy: testvalues[1].policy,
			//}).Return(testvalues[1].policy, nil)
			defer mockCRM.AssertExpectations(t)
			policy, err := client.AddSAtoRole(value.saname, value.roles, value.project, value.condition)
			//should return error
			if test := assert.Errorf(t, err, "\tnot expected: AddSAtoRole do not returned error."); test {
				t.Log("\texpected: AddSAtoRole() returned error.")
				if test := assert.Containsf(t, err.Error(), "When checking if policy was modified for", "\tnot expected: AddSAtoRole() returned unexpected error message."); test {
					t.Log("\texpected: AddSAtoRole() returned expected error message.")
				}
			}
			//should return nil policy
			if test := assert.Nil(t, policy, "\tnotexpected: AddSAtoRole returned not nil policy."); test {
				t.Log("\texpected: AddSAtoRole returned nil policy.")
			}
			//should call crmservice.GetPolicy
			if test := mockCRM.AssertCalled(t, "GetPolicy", value.project, &cloudresourcemanager.GetIamPolicyRequest{}); test {
				t.Log("\texpected: AddSAtoRole() did call crmservice.GetPolicy().")
			} else {
				t.Log("\tnot expected: AddSAtoRole() did not call crmservice.GetPolicy().")
			}
			//should call crmservice.GetPolicy once
			if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 2); test {
				t.Log("\texpected: AddSAtoRole() called crmservice.GetPolicy() expected number of times.")
			} else {
				t.Log("\tnot expected: AddSAtoRole() called crmservice.GetPolicy() unexpected number of times.")
			}
			//should not call crmservice.SetPolicy
			if test := mockCRM.AssertNotCalled(t, "SetPolicy"); test {
				t.Log("\texpected: AddSAtoRole() did not call crmservice.SetPolicy().")
			} else {
				t.Log("\tnot expected: AddSAtoRole() did call crmservice.SetPolicy().")
			}
			//client.policies should not contain project policy
			if test := assert.Nilf(t, client.policies[value.project], "\tnot expected: Client object holds policy for project: %s.", value.project); test {
				t.Logf("\texpected: Client object do not contain policy for project: %s.", value.project)
			}
			//test with correct values
			//should return error
			//should return nil policy
			//should call crmservice.GetPolicy
			//should call crmservice.GetPolicy twice
			//should not call crmservice.SetPolicy
			//client.policies should not contain project policy
		})

		t.Run("AddSAtoRole should fail because got error when setting policy in GCP", func(t *testing.T) {
			//test with correct values
			value := testvalue{saname: "test_sa_01", project: "test_project_01", roles: []string{"test_role_01"}, policy: &cloudresourcemanager.Policy{
				AuditConfigs: nil,
				Bindings: []*cloudresourcemanager.Binding{&cloudresourcemanager.Binding{
					Condition:       nil,
					Members:         []string{"group:prow_admins@sap.com", "serviceAccount:some_sa@test_project.iam.gserviceaccount.com"},
					Role:            "roles/owner",
					ForceSendFields: nil,
					NullFields:      nil,
				}, &cloudresourcemanager.Binding{
					Condition:       nil,
					Members:         []string{"serviceAccount:service-727270599349@gs-project-accounts.iam.gserviceaccount.com", "user:some_user@sap.com"},
					Role:            "roles/compute.admin",
					ForceSendFields: nil,
					NullFields:      nil,
				}},
				Etag:            "",
				Version:         0,
				ServerResponse:  googleapi.ServerResponse{},
				ForceSendFields: nil,
				NullFields:      nil,
			}, condition: nil}
			mockCRM := &mocks.CRM{}
			client, _ := New(mockCRM)
			mockCRM.On("GetPolicy", value.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(, nil)
			//mockCRM.On("SetPolicy", value.project, &cloudresourcemanager.SetIamPolicyRequest{
			//	Policy: testvalues[1].policy,
			//}).Return(testvalues[1].policy, nil)
			defer mockCRM.AssertExpectations(t)
			policy, err := client.AddSAtoRole(value.saname, value.roles, value.project, value.condition)
			//should return error
			if test := assert.Errorf(t, err, "\tnot expected: AddSAtoRole do not returned error."); test {
				t.Log("\texpected: AddSAtoRole() returned error.")
				if test := assert.Containsf(t, err.Error(), "When checking if policy was modified for", "\tnot expected: AddSAtoRole() returned unexpected error message."); test {
					t.Log("\texpected: AddSAtoRole() returned expected error message.")
				}
			}
			//should return nil policy
			if test := assert.Nil(t, policy, "\tnotexpected: AddSAtoRole returned not nil policy."); test {
				t.Log("\texpected: AddSAtoRole returned nil policy.")
			}
			//should call crmservice.GetPolicy
			if test := mockCRM.AssertCalled(t, "GetPolicy", value.project, &cloudresourcemanager.GetIamPolicyRequest{}); test {
				t.Log("\texpected: AddSAtoRole() did call crmservice.GetPolicy().")
			} else {
				t.Log("\tnot expected: AddSAtoRole() did not call crmservice.GetPolicy().")
			}
			//should call crmservice.GetPolicy once
			if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 2); test {
				t.Log("\texpected: AddSAtoRole() called crmservice.GetPolicy() expected number of times.")
			} else {
				t.Log("\tnot expected: AddSAtoRole() called crmservice.GetPolicy() unexpected number of times.")
			}
			//should not call crmservice.SetPolicy
			if test := mockCRM.AssertNotCalled(t, "SetPolicy"); test {
				t.Log("\texpected: AddSAtoRole() did not call crmservice.SetPolicy().")
			} else {
				t.Log("\tnot expected: AddSAtoRole() did call crmservice.SetPolicy().")
			}
			//client.policies should not contain project policy
			if test := assert.Nilf(t, client.policies[value.project], "\tnot expected: Client object holds policy for project: %s.", value.project); test {
				t.Logf("\texpected: Client object do not contain policy for project: %s.", value.project)
			}
			//test with correct values
			//should return error
			//should return nil policy
			//should call crmservice.GetPolicy
			//should call crmservice.GetPolicy twice
			//should call crmservice.SetPolicy
			//should call crmservice.SetPolicy once
			//client.policies should not contain project policy
		})
		t.Run("AddSAtoRole should add serviceaccount to role without errors.", func(t *testing.T) {
			//test with correct values and multiple roles
			//should return nil error
			//should return *cloudresourcemanager.Policy
			//should call crmservice.GetPolicy
			//should call crmservice.GetPolicy twice
			//should call crmservice.SetPolicy
			//should call crmservice.SetPolicy once
			//client.policies should contain project policy with correct binding
			//client.policies should contain project policy with provided member
			//Returned policy should be equal to the client.policies project policy
		})
	*/
}

func TestClient_getPolicy(t *testing.T) {
	t.Run("getPolicy() should get policy without errors.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		mockCRM.On("GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}).Return(testvalues[1].policy, nil)
		defer mockCRM.AssertExpectations(t)
		policy, err := client.getPolicy("test-project")
		require.Nilf(t, err, "\tgetPolicy() returned not nil error.")
		t.Log("\tgetPolicy() returned nil error.")
		require.IsTypef(t, &cloudresourcemanager.Policy{}, policy, "\tgetPolicy() returned policy object not type of cloudresourcemanager.Policy.")
		t.Log("\tgetPolicy() returned policy object of type cloudresourcemanager.Policy.")
		require.ElementsMatchf(t, testvalues[1].policy.Bindings, policy.Bindings, "\tgetPolicy() returned policy with different Bindings slice than expected.")
		t.Log("\tgetPolicy() returned policy with expected Bindings slice.")
		if test := mockCRM.AssertCalled(t, "GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\tcrmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was called with unexpected arguments.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\tcrmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was called unexpected number of times.")
		}
	})
	t.Run("getPolicy() returned http.StatusNotModified.", func(t *testing.T) {

		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		client.policies["test-project"] = testvalues[1].policy
		mockCRM.On("GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, notModifiedError)
		defer mockCRM.AssertExpectations(t)
		policy, err := client.getPolicy("test-project")
		var test bool
		exit := false
		for loop := true; loop; loop = !(exit == true) {
			code := googleapi.IsNotModified(err)
			if code == true {
				test = true
				exit = true
			} else {
				err = errors.Unwrap(err)
				if err == nil {
					exit = true
					test = false
				}
			}
		}
		require.Truef(t, test, "\tgetPolicy() returned error did not contain http.StatusNotModified code.")
		t.Log("\tgetPolicy() returned error containing http.StatusNotModified code.")
		require.Nilf(t, policy, "\tgetPolicy() returned not nil policy object.")
		t.Log("\tgetPolicy() returned nil policy object.")
		if test := mockCRM.AssertCalled(t, "GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\tcrmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was called with unexpected arguments.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\tcrmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was called unexpected number of times.")
		}
	})

	t.Run("getPolicy() should fail and return not nil error.", func(t *testing.T) {
		testError := errors.New("Get test-project policy error")
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		mockCRM.On("GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, testError)
		defer mockCRM.AssertExpectations(t)
		policy, err := client.getPolicy("test-project")
		require.EqualErrorf(t, err, "When downloading policy for test-project project got error: [Get test-project policy error].", "\tgetPolicy() returned error message other than expected.")
		t.Log("\tgetPolicy() returned expected error message.")
		require.Nilf(t, policy, "\tgetPolicy() returned not nil policy object. %s")
		t.Log("\tgetPolicy() returned nil policy object.")
		if test := mockCRM.AssertCalled(t, "GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\tcrmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was called with unexpected arguments.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\tcrmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was called unexpected number of times.")
		}
	})
}

func TestClient_addToRole(t *testing.T) {
	//test with not existing role
	//test with existing role with not matching condition
	t.Run("addToRole should fail because missing role binding.", func(t *testing.T) {
		type testvalues = []struct {
			saname    string
			project   string
			roles     []string
			policy    *cloudresourcemanager.Policy
			condition *cloudresourcemanager.Expr
		}
		values := testvalues{
			{saname: "test_sa_01", project: "test_project_01", roles: []string{"test_role_01"}, policy: &cloudresourcemanager.Policy{
				AuditConfigs: nil,
				Bindings: []*cloudresourcemanager.Binding{&cloudresourcemanager.Binding{
					Condition:       nil,
					Members:         []string{"group:prow_admins@sap.com", "serviceAccount:some_sa@test_project.iam.gserviceaccount.com"},
					Role:            "roles/owner",
					ForceSendFields: nil,
					NullFields:      nil,
				}, &cloudresourcemanager.Binding{
					Condition:       nil,
					Members:         []string{"serviceAccount:service-727270599349@gs-project-accounts.iam.gserviceaccount.com", "user:some_user@sap.com"},
					Role:            "roles/compute.admin",
					ForceSendFields: nil,
					NullFields:      nil,
				}},
				Etag:            "",
				Version:         0,
				ServerResponse:  googleapi.ServerResponse{},
				ForceSendFields: nil,
				NullFields:      nil,
			}, condition: &cloudresourcemanager.Expr{
				Description:     "",
				Expression:      "test-expression",
				Location:        "",
				Title:           "",
				ForceSendFields: nil,
				NullFields:      nil,
			}},
			{saname: "test_sa_04", project: "test_project_04", roles: []string{"roles/test_role_04"}, policy: &cloudresourcemanager.Policy{
				AuditConfigs: nil,
				Bindings: []*cloudresourcemanager.Binding{&cloudresourcemanager.Binding{
					Condition:       nil,
					Members:         []string{"group:prow_admins@sap.com", "serviceAccount:some_sa@test_project.iam.gserviceaccount.com"},
					Role:            "roles/owner",
					ForceSendFields: nil,
					NullFields:      nil,
				}, &cloudresourcemanager.Binding{
					Condition: &cloudresourcemanager.Expr{
						Description:     "",
						Expression:      "test-expression",
						Location:        "",
						Title:           "",
						ForceSendFields: nil,
						NullFields:      nil,
					},
					Members:         []string{"serviceAccount:service-727270599349@gs-project-accounts.iam.gserviceaccount.com", "user:some_user@sap.com"},
					Role:            "roles/test_role_04",
					ForceSendFields: nil,
					NullFields:      nil,
				}},
				Etag:            "",
				Version:         0,
				ServerResponse:  googleapi.ServerResponse{},
				ForceSendFields: nil,
				NullFields:      nil,
			}, condition: nil},
		}
		for _, tv := range values {
			t.Logf("\n\tTesting with values:\n\tsaname: [%s]\n\tproject: [%s]\n\troles: %v\n\tpolicy: [%+v]\n\tcondition: [%+v]", tv.saname, tv.project, tv.roles, tv.policy, tv.condition)
			mockCRM := &mocks.CRM{}
			client, _ := New(mockCRM)
			client.policies[tv.project] = tv.policy
			rolefullname := client.makeRoleFullname(tv.roles[0])
			safqdn := client.makeSafqdn(tv.saname, tv.project)
			err := client.addToRole(safqdn, rolefullname, tv.project, tv.condition)
			bindingpresent := false
			if test := assert.IsTypef(t, &BindingNotFoundError{}, err, "\tnot expected: addToRole() did not returned BindingNotFoundError"); test {
				t.Log("\texpected: addToRole() returned BindingNotFoundError")
			}
			for _, binding := range client.policies[tv.project].Bindings {
				if binding.Role == rolefullname && binding.Condition == tv.condition {
					bindingpresent = true
				}
			}
			if test := assert.Falsef(t, bindingpresent, "\tnot expected: Role binding found in policy."); test {
				t.Log("\texpected: Binding not present in policy.")
			}
		}
	})
	//test with existing role
	t.Run("addToRole should add sa to role without errors.", func(t *testing.T) {
		type testvalues = []struct {
			saname    string
			project   string
			roles     []string
			policy    *cloudresourcemanager.Policy
			condition *cloudresourcemanager.Expr
		}
		values := testvalues{
			{
				saname:  "test_sa_01",
				project: "test_project_01",
				roles:   []string{"organizations/test_role_01"},
				policy: &cloudresourcemanager.Policy{
					AuditConfigs: nil,
					Bindings: []*cloudresourcemanager.Binding{
						&cloudresourcemanager.Binding{
							Condition:       nil,
							Members:         []string{"group:prow_admins@sap.com", "serviceAccount:some_sa@test_project.iam.gserviceaccount.com"},
							Role:            "roles/owner",
							ForceSendFields: nil,
							NullFields:      nil,
						},
						&cloudresourcemanager.Binding{
							Condition: &cloudresourcemanager.Expr{
								Description:     "",
								Expression:      "test-expression",
								Location:        "",
								Title:           "",
								ForceSendFields: nil,
								NullFields:      nil,
							},
							Members:         []string{"serviceAccount:service-727270599349@gs-project-accounts.iam.gserviceaccount.com", "user:some_user@sap.com"},
							Role:            "organizations/test_role_01",
							ForceSendFields: nil,
							NullFields:      nil,
						},
					},
					Etag:            "",
					Version:         0,
					ServerResponse:  googleapi.ServerResponse{},
					ForceSendFields: nil,
					NullFields:      nil,
				},
				condition: &cloudresourcemanager.Expr{
					Description:     "",
					Expression:      "test-expression",
					Location:        "",
					Title:           "",
					ForceSendFields: nil,
					NullFields:      nil,
				},
			},
		}
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		client.policies[values[0].project] = values[0].policy
		rolefullname := client.makeRoleFullname(values[0].roles[0])
		safqdn := client.makeSafqdn(values[0].saname, values[0].project)
		err := client.addToRole(safqdn, rolefullname, values[0].project, values[0].condition)
		bindingpresent := false
		require.Nilf(t, err, "\tnot expected: addToRole() returned not nil error")
		t.Log("\texpected: addToRole() returned nil error")
		for _, binding := range client.policies[values[0].project].Bindings {
			if binding.Role == rolefullname && cmp.Equal(binding.Condition, values[0].condition) {
				bindingpresent = true
				require.Containsf(t, binding.Members, safqdn, "\tnot expected: Correct binding do not contain expected member.")
				t.Log("\texpected: Correct binding contain expected member.")
			}
		}
		if test := assert.Truef(t, bindingpresent, "\tnot expected: Correct role binding not found in policy."); test {
			t.Log("\texpected: Correct role binding present in policy.")
		}
	})
}

func TestClient_makeSafqdn(t *testing.T) {
	t.Run("makeSafqdn() should return GCP policy valid FQDN serviceaccount name.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		safqdn := client.makeSafqdn(testvalues[0].saname, testvalues[0].project)
		require.Equalf(t, fmt.Sprintf("serviceAccount:%s@%s.iam.gserviceaccount.com", testvalues[0].saname, testvalues[0].project), safqdn, "\tmakeSafqdn() returned unexpected string value.")
		t.Log("\tmakeSafqdn() returned expected string value.")
	})
}

//Check which entries from testvalues to use.
func TestClient_makeRoleFullname(t *testing.T) {
	for _, tv := range testvalues {
		t.Logf("Testing with values:\n\trole: %v", tv.roles)
		var role string
		if prefixed, _ := regexp.MatchString("^(roles|organizations)/.*", tv.roles[0]); prefixed {
			role = tv.roles[0]
		} else {
			role = fmt.Sprintf("roles/%s", tv.roles[0])
		}
		t.Run("makeRoleFullname() shuold return GCP policy valid name", func(t *testing.T) {
			mockCRM := &mocks.CRM{}
			client, _ := New(mockCRM)
			rolefullname := client.makeRoleFullname(tv.roles[0])
			assert.Equalf(t, role, rolefullname, "\tmakeRoleFullname() returned unexpected value for GCP policy role name.")
			t.Log("\tmakeRoleFulllname() returned expected GCP policy role name.")
		})
	}
}

func TestClient_addRole(t *testing.T) {
	for _, tv := range testvalues[:1] {
		t.Run("addRole() added expected role to the policy.", func(t *testing.T) {
			mockCRM := mocks.CRM{}
			client, _ := New(&mockCRM)
			client.policies[tv.project] = tv.policy
			safqdn := client.makeSafqdn(tv.saname, tv.project)
			rolefullname := client.makeRoleFullname(tv.roles[0])
			client.addRole(safqdn, rolefullname, tv.project, tv.condition)
			rolepresent := false
			for _, binding := range client.policies[tv.project].Bindings {
				assert.IsTypef(t, &cloudresourcemanager.Binding{}, binding, "\tPolicy contain binding which is not type of *cloudresourcemanager.Binding [%+v].", binding)
				t.Log("'tAll policy bindings are of expected type.")
				if binding.Role == rolefullname {
					rolepresent = true
					require.Lenf(t, binding.Members, 1, "\tAdded binding do not contain expected number of members.")
					t.Log("\tAdded binding contain expected number of members.")
					require.Containsf(t, binding.Members, safqdn, "\tAdded binding do not contain expected member.")
					t.Log("\tAdded binding contain expected member.")
					require.Equalf(t, binding.Condition, tv.condition, "\tAdded binding do not contain expected condition.")
					t.Log("\tAdded binding contain expected condition.")
				}
			}
			require.Truef(t, rolepresent, "\tBinding for added role is not present in policy.")
			t.Log("\tBinding for added role is present in policy.")
		})
	}
}

func TestClient_setPolicy(t *testing.T) {
	notModifiedError := &googleapi.Error{
		Code:    http.StatusNotModified,
		Message: "",
		Body:    "",
		Header:  nil,
		Errors:  nil,
	}
	projectname := "test-project"
	t.Run("setPolicy() set policy on GCP without errors.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		client.policies[projectname] = testvalues[1].policy
		mockCRM.On("GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, notModifiedError)
		mockCRM.On("SetPolicy", projectname, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: testvalues[1].policy,
		}).Return(testvalues[1].policy, nil)
		defer mockCRM.AssertExpectations(t)
		policy, err := client.setPolicy(projectname)
		assert.IsTypef(t, &cloudresourcemanager.Policy{}, policy, "\tsetPolicy() returned policy is not expected type.")
		t.Log("\tsetPolicy() returned policy of expected type.")
		assert.Equalf(t, testvalues[1].policy, policy, "\tsetPolicy() returned policy is not same as provided.")
		t.Log("\tsetPolicy() returned equal policy as provided.")
		assert.Nilf(t, err, "\tsetPolicy() returned not nil error")
		t.Log("\tsetPolicy() returned nil error")
		if test := mockCRM.AssertCalled(t, "GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\tcrmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was not called.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\tcrmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was called unexpected number of times.")
		}
		if test := mockCRM.AssertCalled(t, "SetPolicy", projectname, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: testvalues[1].policy,
		}); test {
			t.Log("\tcrmservice.SetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tcrmservice.SetPolicy() was not called.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "SetPolicy", 1); test {
			t.Log("\tcrmservice.SetPolicy() was called expected number of times.")
		} else {
			t.Log("\tcrmservice.SetPolicy() was called unexpected number of times.")
		}
	})
	t.Run("setPolicy() returned PolicyModifiedError error.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		client.policies[projectname] = testvalues[1].policy
		client.policiesetag[projectname] = "TestEtag"
		mockCRM.On("GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}).Return(testvalues[1].policy, nil)
		defer mockCRM.AssertExpectations(t)
		policy, err := client.setPolicy(projectname)
		if test := assert.Nilf(t, policy, "\tnot expected: setPolicy() returned not nil policy."); test {
			t.Log("\texpected: setPolicy() returned nil policy.")
		}
		if test := assert.Errorf(t, err, "\tnot expected: setPolicy() did not returned PolicyModifiedError"); test {
			t.Log("\texpected: setPolicy() returned PolicyModifiedError")
		}
		if test := mockCRM.AssertCalled(t, "GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\tcrmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was not called.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\tcrmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was called unexpected number of times.")
		}
		if test := mockCRM.AssertNotCalled(t, "SetPolicy", projectname, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: testvalues[1].policy,
		}); test {
			t.Log("\texpected: crmservice.SetPolicy() was not called.")
		} else {
			t.Log("\tnot expected: crmservice.SetPolicy() was called.")
		}
	})
	t.Run("setPolicy() returned non PolicyModifiedError error when checking if policy was modified.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		client.policies[projectname] = testvalues[1].policy
		mockCRM.On("GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, errors.New("test-error"))
		defer mockCRM.AssertExpectations(t)
		policy, err := client.setPolicy(projectname)
		if test := assert.Nilf(t, policy, "\tnot expected: setPolicy() returned not nil policy."); test {
			t.Log("\texpected: setPolicy() returned nil policy")
		}
		if test := assert.NotNilf(t, err, "\tnot expected: setPolicy() returned nil error."); test {
			t.Log("\tsetPolicy() returned not nil error")
		}
		if test := mockCRM.AssertCalled(t, "GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\tcrmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was not called.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\tcrmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was called unexpected number of times.")
		}
		if test := mockCRM.AssertNotCalled(t, "SetPolicy", projectname, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: testvalues[1].policy,
		}); test {
			t.Log("\texpected: crmservice.SetPolicy() was not called.")
		} else {
			t.Log("\tnot expected: crmservice.SetPolicy() was called.")
		}
	})
	t.Run("setPolicy() returned error when setting new policy.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		client.policies[projectname] = testvalues[1].policy
		mockCRM.On("GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, notModifiedError)
		mockCRM.On("SetPolicy", projectname, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: testvalues[1].policy,
		}).Return(nil, errors.New("test-error"))
		defer mockCRM.AssertExpectations(t)
		policy, err := client.setPolicy(projectname)
		if test := assert.Nilf(t, policy, "\tnot expected: setPolicy() returned not nil policy."); test {
			t.Log("\texpected: setPolicy() returned nil policy")
		}
		if test := assert.Errorf(t, err, "\tnot expected: setPolicy() did not returned error."); test {
			t.Log("\tsetPolicy() returned error")
		}
		if test := mockCRM.AssertCalled(t, "GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\tcrmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was not called.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\tcrmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tcrmservice.GetPolicy() was called unexpected number of times.")
		}
		if test := mockCRM.AssertCalled(t, "SetPolicy", projectname, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: testvalues[1].policy,
		}); test {
			t.Log("\texpected: crmservice.SetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tnot expected: crmservice.SetPolicy() was not called with expected arguments.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "SetPolicy", 1); test {
			t.Log("\tcrmservice.SetPolicy() was called expected number of times.")
		} else {
			t.Log("\tcrmservice.SetPolicy() was called unexpected number of times.")
		}
	})
}
