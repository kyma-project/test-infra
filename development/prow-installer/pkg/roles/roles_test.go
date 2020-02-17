package roles

import (
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/roles/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
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
		testvalues := []testvalue{
			{saname: "", project: "test_project_01", roles: []string{"test_role_01"}, policy: nil, condition: nil},
			{saname: "test_sa_01", project: "", roles: []string{"test_role_01"}, policy: nil, condition: nil},
			{saname: "test_sa_01", project: "test_project_01", roles: nil, policy: nil, condition: nil},
			{saname: "test_sa_01", project: "test_project_01", roles: []string{""}, policy: nil, condition: nil},
		}
		for _, tv := range testvalues {
			mockCRM := &mocks.CRM{}
			client, _ := New(mockCRM)

			policy, err := client.AddSAtoRole(tv.saname, tv.roles, tv.project, tv.condition)
			//should return error
			if test := assert.EqualErrorf(t, err, fmt.Sprintf("One of mandatory method arguments saname, projectname ,role can not be empty. Got values. saname: %s projectname: %s roles: %v.", tv.saname, tv.project, tv.roles), "\tnot expected: AddSAtoRole returned unexpected error or nil."); test {
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
		}
	})

	t.Run("AddSAtoRole should fail because safqdn passed as saname argument", func(t *testing.T) {
		//test with safqdn passed as saname
		tv := testvalue{
			saname:    "test_sa_01@test_project_01.iam.gserviceaccount.com",
			project:   "test_project_01",
			roles:     []string{"test_role_01"},
			policy:    nil,
			condition: nil,
		}
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)

		policy, err := client.AddSAtoRole(tv.saname, tv.roles, tv.project, tv.condition)
		//should return error
		if test := assert.EqualErrorf(t, err, fmt.Sprintf("saname argument can not be serviceaccount fqdn. Provide only name, without domain part. Got value: %s.", tv.saname), "\tnot expected: AddSAtoRole returned unexpected error or nil."); test {
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
	})

	t.Run("AddSAtoRole should fail because got error when getting policy from GCP", func(t *testing.T) {
		//test with correct arguments
		tv := testvalue{
			saname:  "test_sa_01",
			project: "test_project_01",
			roles:   []string{"test_role_01"},
			policy: &cloudresourcemanager.Policy{
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
			},
			condition: nil}

		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)

		mockCRM.On("GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, errors.New("GetPolicy() error."))
		defer mockCRM.AssertExpectations(t)

		policy, err := client.AddSAtoRole(tv.saname, tv.roles, tv.project, tv.condition)
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
		if test := mockCRM.AssertCalled(t, "GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}); test {
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
	})

	t.Run("AddSAtoRole should fail because got PolicyModifiedError when setting policy in GCP", func(t *testing.T) {
		//test with correct values
		tv := testvalue{
			saname:  "test_sa_01",
			project: "test_project_01",
			roles:   []string{"test_role_01"},
			policy: &cloudresourcemanager.Policy{
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
				Etag:            "initial-Etag",
				Version:         0,
				ServerResponse:  googleapi.ServerResponse{},
				ForceSendFields: nil,
				NullFields:      nil,
			},
			condition: nil}

		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)

		mockCRM.On("GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(func(string, *cloudresourcemanager.GetIamPolicyRequest) *cloudresourcemanager.Policy {
			if len(mockCRM.Calls) == 1 {
				return tv.policy
			} else if len(mockCRM.Calls) == 2 {
				return &cloudresourcemanager.Policy{Etag: "different-Etag"}
			}
			return nil
		}, nil)
		defer mockCRM.AssertExpectations(t)

		policy, err := client.AddSAtoRole(tv.saname, tv.roles, tv.project, tv.condition)
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
		if test := mockCRM.AssertCalled(t, "GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}); test {
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
	})

	t.Run("AddSAtoRole should fail because got error when setting policy in GCP", func(t *testing.T) {
		//test with correct values
		tv := testvalue{
			saname:  "test_sa_01",
			project: "test_project_01",
			roles:   []string{"test_role_01"},
			policy: &cloudresourcemanager.Policy{
				AuditConfigs: nil,
				Bindings: []*cloudresourcemanager.Binding{
					&cloudresourcemanager.Binding{
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
				Etag:            "test-Etag",
				Version:         0,
				ServerResponse:  googleapi.ServerResponse{},
				ForceSendFields: nil,
				NullFields:      nil,
			},
			condition: nil,
		}
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		mockCRM.On("GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(tv.policy, nil)
		mockCRM.On("SetPolicy", tv.project, mock.AnythingOfType("*cloudresourcemanager.SetIamPolicyRequest")).Return(nil, errors.New("crmservice.SetPolicy-error"))
		defer mockCRM.AssertExpectations(t)

		policy, err := client.AddSAtoRole(tv.saname, tv.roles, tv.project, tv.condition)
		//should return error
		if test := assert.Errorf(t, err, "\tnot expected: AddSAtoRole did not returned error."); test {
			t.Log("\texpected: AddSAtoRole() returned error.")
			if test := assert.Containsf(t, err.Error(), "When setting new policy for", "\tnot expected: AddSAtoRole() returned unexpected error message."); test {
				t.Log("\texpected: AddSAtoRole() returned expected error message.")
			}
		}
		//should return nil policy
		if test := assert.Nil(t, policy, "\tnotexpected: AddSAtoRole returned not nil policy."); test {
			t.Log("\texpected: AddSAtoRole returned nil policy.")
		}
		//should call crmservice.GetPolicy
		if test := mockCRM.AssertCalled(t, "GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}); test {
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
		if test := mockCRM.AssertCalled(t, "SetPolicy", tv.project, mockCRM.Calls[2].Arguments.Get(1).(*cloudresourcemanager.SetIamPolicyRequest)); test {
			t.Log("\texpected: AddSAtoRole() did call crmservice.SetPolicy().")
		} else {
			t.Log("\tnot expected: AddSAtoRole() did not call crmservice.SetPolicy().")
		}
		//should call crmservice.SetPolicy once
		if test := mockCRM.AssertNumberOfCalls(t, "SetPolicy", 1); test {
			t.Log("\texpected: AddSAtoRole() called crmservice.SetPolicy() expected number of times.")
		} else {
			t.Log("\tnot expected: AddSAtoRole() called crmservice.SetPolicy() unexpected number of times.")
		}
	})


	t.Run("AddSAtoRole should add role without errors.", func(t *testing.T) {
		tv := testvalue{
			saname:  "test_sa_01",
			project: "test_project_01",
			roles:   []string{"test_role_01"},
			policy: &cloudresourcemanager.Policy{
				AuditConfigs: nil,
				Bindings: []*cloudresourcemanager.Binding{
					&cloudresourcemanager.Binding{
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
				Etag:            "test-Etag",
				Version:         0,
				ServerResponse:  googleapi.ServerResponse{},
				ForceSendFields: nil,
				NullFields:      nil,
			},
			condition: nil,
		}
		returnpolicy := &cloudresourcemanager.Policy{
			AuditConfigs:    nil,
			Bindings:        []*cloudresourcemanager.Binding{
				&cloudresourcemanager.Binding{
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
				}, &cloudresourcemanager.Binding{
					Condition:       nil,
					Members:         []string{"serviceAccount:test_sa_01@test_project_01.iam.gserviceaccount.com"},
					Role:            "roles/test_role_01",
					ForceSendFields: nil,
					NullFields:      nil,
			}},
			Etag:            "test-Etag",
			Version:         0,
			ServerResponse:  googleapi.ServerResponse{},
			ForceSendFields: nil,
			NullFields:      nil,
		}
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		mockCRM.On("GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(tv.policy, nil)
		mockCRM.On("SetPolicy", tv.project, &cloudresourcemanager.SetIamPolicyRequest{Policy: returnpolicy}).Return(returnpolicy,nil)
		defer mockCRM.AssertExpectations(t)

		policy, err := client.AddSAtoRole(tv.saname, tv.roles, tv.project, tv.condition)
		//should return error
		if test := assert.Nil(t, err, "\tnot expected: AddSAtoRole returned not nil error."); test {
			t.Log("\texpected: AddSAtoRole() returned nil error.")
		}
		//should return not nil policy
		if test := assert.NotNil(t, policy, "\tnot expected: AddSAtoRole returned nil policy object."); test {
			t.Log("\texpected: AddSAtoRole returned not nil policy object.")
		}
		//should return policy of type *cloudresourcemanager.Policy
		if test := assert.IsType(t, &cloudresourcemanager.Policy{}, policy, "\tnotexpected: AddSAtoRole returned object not type of *cloudresourcemanager.Policy."); test {
			t.Log("\texpected: AddSAtoRole returned object of expected type.")
		}
		//should return same policy as returned by GCP API call
		if test := assert.Equal(t, returnpolicy, policy, "\tnot expected: AddSAtoRole returned different policy than returned by GCP API call."); test {
			t.Log("\texpected: AddSAtoRole returned same policy as GCP API call.")
		}
		//should call crmservice.GetPolicy
		if test := mockCRM.AssertCalled(t, "GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\texpected: AddSAtoRole() did call crmservice.GetPolicy().")
		} else {
			t.Log("\tnot expected: AddSAtoRole() did not call crmservice.GetPolicy().")
		}
		//should call crmservice.GetPolicy twice
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 2); test {
			t.Log("\texpected: AddSAtoRole() called crmservice.GetPolicy() expected number of times.")
		} else {
			t.Log("\tnot expected: AddSAtoRole() called crmservice.GetPolicy() unexpected number of times.")
		}
		//should call crmservice.SetPolicy
		if test := mockCRM.AssertCalled(t, "SetPolicy", tv.project, &cloudresourcemanager.SetIamPolicyRequest{Policy: returnpolicy}); test {
			t.Log("\texpected: AddSAtoRole() did call crmservice.SetPolicy().")
		} else {
			t.Log("\tnot expected: AddSAtoRole() did not call crmservice.SetPolicy().")
		}
		//should call crmservice.SetPolicy once
		if test := mockCRM.AssertNumberOfCalls(t, "SetPolicy", 1); test {
			t.Log("\texpected: AddSAtoRole() called crmservice.SetPolicy() expected number of times.")
		} else {
			t.Log("\tnot expected: AddSAtoRole() called crmservice.SetPolicy() unexpected number of times.")
		}
	})


	t.Run("AddSAtoRole should add serviceaccount to role without errors.", func(t *testing.T) {
		tv := testvalue{
			saname:  "test_sa_01",
			project: "test_project_01",
			roles:   []string{"test_role_01"},
			policy: &cloudresourcemanager.Policy{
				AuditConfigs:    nil,
				Bindings:        []*cloudresourcemanager.Binding{
					&cloudresourcemanager.Binding{
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
					}, &cloudresourcemanager.Binding{
						Condition:       nil,
						Members:         []string{"serviceAccount:some_sa@test_project_01.iam.gserviceaccount.com"},
						Role:            "roles/test_role_01",
						ForceSendFields: nil,
						NullFields:      nil,
					}},
				Etag:            "test-Etag",
				Version:         0,
				ServerResponse:  googleapi.ServerResponse{},
				ForceSendFields: nil,
				NullFields:      nil,
			},
			condition: nil,
		}
		returnpolicy := &cloudresourcemanager.Policy{
			AuditConfigs:    nil,
			Bindings:        []*cloudresourcemanager.Binding{
				&cloudresourcemanager.Binding{
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
				}, &cloudresourcemanager.Binding{
					Condition:       nil,
					Members:         []string{"serviceAccount:some_sa@test_project_01.iam.gserviceaccount.com", "serviceAccount:test_sa_01@test_project_01.iam.gserviceaccount.com"},
					Role:            "roles/test_role_01",
					ForceSendFields: nil,
					NullFields:      nil,
				}},
			Etag:            "test-Etag",
			Version:         0,
			ServerResponse:  googleapi.ServerResponse{},
			ForceSendFields: nil,
			NullFields:      nil,
		}
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		mockCRM.On("GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(tv.policy, nil)
		mockCRM.On("SetPolicy", tv.project, &cloudresourcemanager.SetIamPolicyRequest{Policy: returnpolicy}).Return(returnpolicy,nil)
		defer mockCRM.AssertExpectations(t)

		policy, err := client.AddSAtoRole(tv.saname, tv.roles, tv.project, tv.condition)
		//should return error
		if test := assert.Nil(t, err, "\tnot expected: AddSAtoRole returned not nil error."); test {
			t.Log("\texpected: AddSAtoRole() returned nil error.")
		}
		//should return not nil policy
		if test := assert.NotNil(t, policy, "\tnot expected: AddSAtoRole returned nil policy object."); test {
			t.Log("\texpected: AddSAtoRole returned not nil policy object.")
		}
		//should return policy of type *cloudresourcemanager.Policy
		if test := assert.IsType(t, &cloudresourcemanager.Policy{}, policy, "\tnotexpected: AddSAtoRole returned object not type of *cloudresourcemanager.Policy."); test {
			t.Log("\texpected: AddSAtoRole returned object of expected type.")
		}
		//should return same policy as returned by GCP API call
		if test := assert.Equal(t, returnpolicy, policy, "\tnot expected: AddSAtoRole returned different policy than returned by GCP API call."); test {
			t.Log("\texpected: AddSAtoRole returned same policy as GCP API call.")
		}
		//should call crmservice.GetPolicy
		if test := mockCRM.AssertCalled(t, "GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\texpected: AddSAtoRole() did call crmservice.GetPolicy().")
		} else {
			t.Log("\tnot expected: AddSAtoRole() did not call crmservice.GetPolicy().")
		}
		//should call crmservice.GetPolicy twice
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 2); test {
			t.Log("\texpected: AddSAtoRole() called crmservice.GetPolicy() expected number of times.")
		} else {
			t.Log("\tnot expected: AddSAtoRole() called crmservice.GetPolicy() unexpected number of times.")
		}
		//should call crmservice.SetPolicy
		if test := mockCRM.AssertCalled(t, "SetPolicy", tv.project, &cloudresourcemanager.SetIamPolicyRequest{Policy: returnpolicy}); test {
			t.Log("\texpected: AddSAtoRole() did call crmservice.SetPolicy().")
		} else {
			t.Log("\tnot expected: AddSAtoRole() did not call crmservice.SetPolicy().")
		}
		//should call crmservice.SetPolicy once
		if test := mockCRM.AssertNumberOfCalls(t, "SetPolicy", 1); test {
			t.Log("\texpected: AddSAtoRole() called crmservice.SetPolicy() expected number of times.")
		} else {
			t.Log("\tnot expected: AddSAtoRole() called crmservice.SetPolicy() unexpected number of times.")
		}
	})

}

func TestClient_getPolicy(t *testing.T) {
	type testvalue = struct {
		saname    string
		project   string
		roles     []string
		policy    *cloudresourcemanager.Policy
		condition *cloudresourcemanager.Expr
	}

	testvalues := []testvalue{
		{
			saname:  "test_sa_02",
			project: "test_project_02",
			roles:   []string{"test_role_02"},
			policy: &cloudresourcemanager.Policy{
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
				},
				},
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
	}

	t.Run("getPolicy() should get policy without errors.", func(t *testing.T) {
		tv := testvalues[0]
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)

		mockCRM.On("GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(tv.policy, nil)
		defer mockCRM.AssertExpectations(t)

		policy, err := client.getPolicy(tv.project)

		if test := assert.Nilf(t, err, "\tnot expected: getPolicy() returned not nil error."); test {
			t.Log("\texpected: getPolicy() returned nil error.")
		}

		if test := assert.IsTypef(t, &cloudresourcemanager.Policy{}, policy, "\tnot expected: getPolicy() returned policy object not type of cloudresourcemanager.Policy."); test {
			t.Log("\texpected: getPolicy() returned policy object of type cloudresourcemanager.Policy.")
		}

		if test := assert.ElementsMatchf(t, tv.policy.Bindings, policy.Bindings, "\tnot expected: getPolicy() returned policy with different Bindings slice than expected."); test {
			t.Log("\texpected: getPolicy() returned policy with expected Bindings slice.")
		}
		if test := mockCRM.AssertCalled(t, "GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\texpected: crmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tnot expected: crmservice.GetPolicy() was called with unexpected arguments.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\texpected: crmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tnot expected: crmservice.GetPolicy() was called unexpected number of times.")
		}
	})

	t.Run("getPolicy() should fail and return not nil error.", func(t *testing.T) {
		testError := errors.New("Get test-project policy error")
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)

		mockCRM.On("GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, testError)
		defer mockCRM.AssertExpectations(t)

		policy, err := client.getPolicy("test-project")

		if test := assert.Errorf(t, err, "\tgetPolicy() returned error message other than expected."); test {
			t.Log("\tgetPolicy() returned expected error message.")
		}
		if test := assert.Nilf(t, policy, "\tgetPolicy() returned not nil policy object. %s"); test {
			t.Log("\tgetPolicy() returned nil policy object.")
		}
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
	type testvalue = struct {
		saname    string
		project   string
		roles     []string
		policy    *cloudresourcemanager.Policy
		condition *cloudresourcemanager.Expr
	}
	testvalues := []testvalue{
		{
			saname:  "test_sa_01",
			project: "test_project_01",
			roles:   []string{"test_role_01"},
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
		{
			saname:  "test_sa_04",
			project: "test_project_04",
			roles:   []string{"test_role_04"},
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
						Role:            "roles/test_role_04",
						ForceSendFields: nil,
						NullFields:      nil,
					}},
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
	//test with not existing role
	//test with existing role with not matching condition
	t.Run("addToRole should fail because missing role binding.", func(t *testing.T) {
		tv := testvalues[0]
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		policy := tv.policy
		rolefullname := client.makeRoleFullname(tv.roles[0])
		safqdn := client.makeSafqdn(tv.saname, tv.project)
		err := client.addToRole(policy, safqdn, rolefullname, tv.project, tv.condition)
		bindingpresent := false
		if test := assert.IsTypef(t, &BindingNotFoundError{}, err, "\tnot expected: addToRole() did not returned BindingNotFoundError"); test {
			t.Log("\texpected: addToRole() returned BindingNotFoundError")
		}
		for _, binding := range policy.Bindings {
			if binding.Role == rolefullname && cmp.Equal(binding.Condition, tv.condition) {
				bindingpresent = true
			}
		}
		if test := assert.Falsef(t, bindingpresent, "\tnot expected: Role binding found in policy."); test {
			t.Log("\texpected: Binding not present in policy.")
		}
	})
	//test with existing role
	t.Run("addToRole should add sa to role without errors.", func(t *testing.T) {
		tv := testvalues[1]
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		policy := tv.policy
		rolefullname := client.makeRoleFullname(tv.roles[0])
		safqdn := client.makeSafqdn(tv.saname, tv.project)
		err := client.addToRole(policy, safqdn, rolefullname, tv.project, tv.condition)
		bindingpresent := false
		if test := assert.Nilf(t, err, "\tnot expected: addToRole() returned not nil error"); test {
			t.Log("\texpected: addToRole() returned nil error")
		}
		for _, binding := range policy.Bindings {
			if binding.Role == rolefullname && cmp.Equal(binding.Condition, tv.condition) {
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
	type testvalue struct {
		saname  string
		project string
		safqdn  string
	}

	testvalues := []testvalue{
		{saname: "test-sa-01", project: "test-project-01", safqdn: "serviceAccount:test-sa-01@test-project-01.iam.gserviceaccount.com"},
	}

	t.Run("makeSafqdn() should return GCP policy valid FQDN serviceaccount name.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		safqdn := client.makeSafqdn(testvalues[0].saname, testvalues[0].project)
		require.Equalf(t, testvalues[0].safqdn, safqdn, "\tnot expected: makeSafqdn() returned unexpected string value.")
		t.Log("\texpected: makeSafqdn() returned expected string value.")
	})
}

//Check which entries from testvalues to use.
func TestClient_makeRoleFullname(t *testing.T) {
	type testvalue = struct {
		role         string
		rolefullname string
	}
	testvalues := []testvalue{
		{role: "test-role-01", rolefullname: "roles/test-role-01"},
		{role: "organizations/test-role-02", rolefullname: "organizations/test-role-02"},
		{role: "roles/test-role-03", rolefullname: "roles/test-role-03"},
	}

	for _, tv := range testvalues {
		t.Run("makeRoleFullname() should return GCP policy valid name", func(t *testing.T) {
			mockCRM := &mocks.CRM{}
			client, _ := New(mockCRM)
			rolefullname := client.makeRoleFullname(tv.role)
			assert.Equalf(t, tv.rolefullname, rolefullname, "\tnot expected: makeRoleFullname() returned unexpected value for GCP policy role name.")
			t.Log("\texpected: makeRoleFulllname() returned expected GCP policy role name.")
		})
	}
}

func TestClient_addRole(t *testing.T) {
	type testvalue = struct {
		saname    string
		project   string
		roles     []string
		policy    *cloudresourcemanager.Policy
		condition *cloudresourcemanager.Expr
	}

	testvalues := []testvalue{
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
	}

	for _, tv := range testvalues[:1] {
		t.Run("addRole() added expected role to the policy.", func(t *testing.T) {
			mockCRM := mocks.CRM{}
			client, _ := New(&mockCRM)
			policy := tv.policy
			safqdn := client.makeSafqdn(tv.saname, tv.project)
			rolefullname := client.makeRoleFullname(tv.roles[0])

			client.addRole(policy, safqdn, rolefullname, tv.project, tv.condition)
			rolepresent := false
			for _, binding := range policy.Bindings {
				assert.IsTypef(t, &cloudresourcemanager.Binding{}, binding, "\tnot expected: Policy contain binding which is not type of *cloudresourcemanager.Binding [%+v].", binding)
				t.Log("\texpected: All policy bindings are of expected type.")
				if binding.Role == rolefullname {
					rolepresent = true
					require.Lenf(t, binding.Members, 1, "\tnot expected: Added binding do not contain expected number of members.")
					t.Log("\texpected: Added binding contain expected number of members.")
					require.Containsf(t, binding.Members, safqdn, "\tnot expected: Added binding do not contain expected member.")
					t.Log("\texpected: Added binding contain expected member.")
					require.Equalf(t, binding.Condition, tv.condition, "\tnot expected: Added binding do not contain expected condition.")
					t.Log("\texpected: Added binding contain expected condition.")
				}
			}
			require.Truef(t, rolepresent, "\tBinding for added role is not present in policy.")
			t.Log("\tBinding for added role is present in policy.")
		})
	}
}

func TestClient_setPolicy(t *testing.T) {
	type testvalue = struct {
		saname    string
		project   string
		roles     []string
		policy    *cloudresourcemanager.Policy
		condition *cloudresourcemanager.Expr
	}

	testvalues := []testvalue{
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
	}

	projectname := "test-project"
	t.Run("setPolicy() set policy on GCP without errors.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		policy := testvalues[0].policy
		projectname := testvalues[0].project

		mockCRM.On("GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}).Return(testvalues[0].policy, nil)
		mockCRM.On("SetPolicy", projectname, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: testvalues[0].policy,
		}).Return(testvalues[0].policy, nil)
		defer mockCRM.AssertExpectations(t)

		err := client.setPolicy(policy, projectname)

		assert.IsTypef(t, &cloudresourcemanager.Policy{}, policy, "\tnot expected: setPolicy() returned policy is not expected type.")
		t.Log("\texpected: setPolicy() returned policy of expected type.")

		assert.Equalf(t, testvalues[0].policy, policy, "\tnot expected: setPolicy() returned policy is not same as provided.")
		t.Log("\texpected: setPolicy() returned equal policy as provided.")

		assert.Nilf(t, err, "\tnot expected: setPolicy() returned not nil error")
		t.Log("\texpected: setPolicy() returned nil error")

		if test := mockCRM.AssertCalled(t, "GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\texpected: crmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tnot expected: crmservice.GetPolicy() was not called.")
		}

		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\texpected: crmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tnot expected: crmservice.GetPolicy() was called unexpected number of times.")
		}

		if test := mockCRM.AssertCalled(t, "SetPolicy", projectname, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: testvalues[0].policy,
		}); test {
			t.Log("\texpected: crmservice.SetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tnot expected: crmservice.SetPolicy() was not called.")
		}

		if test := mockCRM.AssertNumberOfCalls(t, "SetPolicy", 1); test {
			t.Log("\texpected: crmservice.SetPolicy() was called expected number of times.")
		} else {
			t.Log("\tnot expected: crmservice.SetPolicy() was called unexpected number of times.")
		}
	})

	t.Run("setPolicy() returned PolicyModifiedError error.", func(t *testing.T) {
		tv := testvalues[0]
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		policy := tv.policy

		mockCRM.On("GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}).Return(&cloudresourcemanager.Policy{Etag: "different-Etag"}, nil)
		defer mockCRM.AssertExpectations(t)

		err := client.setPolicy(policy, tv.project)
		if test := assert.IsTypef(t, &PolicyModifiedError{}, err, "\tnot expected: setPolicy() did not returned PolicyModifiedError"); test {
			t.Log("\texpected: setPolicy() returned PolicyModifiedError")
		}
		if test := mockCRM.AssertCalled(t, "GetPolicy", tv.project, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\texpected: crmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tnot expected: crmservice.GetPolicy() was not called.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\texpected: crmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tnot expected: crmservice.GetPolicy() was called unexpected number of times.")
		}
		if test := mockCRM.AssertNotCalled(t, "SetPolicy", tv.project, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: tv.policy,
		}); test {
			t.Log("\texpected: crmservice.SetPolicy() was not called.")
		} else {
			t.Log("\tnot expected: crmservice.SetPolicy() was called.")
		}
	})

	t.Run("setPolicy() returned non PolicyModifiedError error when checking if policy was modified.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		policy := testvalues[0].policy
		projectname := testvalues[0].project
		var e *PolicyModifiedError

		mockCRM.On("GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, errors.New("test-error"))
		defer mockCRM.AssertExpectations(t)

		err := client.setPolicy(policy, projectname)
		if test := assert.Errorf(t, err, "\tnot expected: setPolicy() returned nil error."); test {
			t.Log("\texpected: setPolicy() returned error")
			if !errors.As(err, &e) {
				t.Log("\texpected: setPolicy() returned error not type of PolicyModifiedError")
			} else {
				assert.Fail(t, "\tnot expected: setPolicy() returned error type of PolicyModifiedError")
			}
		}
		if test := mockCRM.AssertCalled(t, "GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\texpected: crmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tnot expected: crmservice.GetPolicy() was not called.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\texpected: crmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tnot expected: crmservice.GetPolicy() was called unexpected number of times.")
		}
		if test := mockCRM.AssertNotCalled(t, "SetPolicy", projectname, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: testvalues[0].policy,
		}); test {
			t.Log("\texpected: crmservice.SetPolicy() was not called.")
		} else {
			t.Log("\tnot expected: crmservice.SetPolicy() was called.")
		}
	})
	t.Run("setPolicy() returned error when setting new policy.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		policy := testvalues[0].policy

		mockCRM.On("GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}).Return(testvalues[0].policy, nil)
		mockCRM.On("SetPolicy", projectname, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: testvalues[0].policy,
		}).Return(nil, errors.New("test-error"))
		defer mockCRM.AssertExpectations(t)

		err := client.setPolicy(policy, projectname)
		if test := assert.Errorf(t, err, "\tnot expected: setPolicy() did not returned error."); test {
			t.Log("\texpected: setPolicy() returned error")
		}
		if test := mockCRM.AssertCalled(t, "GetPolicy", projectname, &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\texpected: crmservice.GetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tnot expected: crmservice.GetPolicy() was not called.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\texpected: crmservice.GetPolicy() was called expected number of times.")
		} else {
			t.Log("\tnot expected: crmservice.GetPolicy() was called unexpected number of times.")
		}
		if test := mockCRM.AssertCalled(t, "SetPolicy", projectname, &cloudresourcemanager.SetIamPolicyRequest{
			Policy: testvalues[0].policy,
		}); test {
			t.Log("\texpected: crmservice.SetPolicy() was called with expected arguments.")
		} else {
			t.Log("\tnot expected: crmservice.SetPolicy() was not called with expected arguments.")
		}
		if test := mockCRM.AssertNumberOfCalls(t, "SetPolicy", 1); test {
			t.Log("\texpected: crmservice.SetPolicy() was called expected number of times.")
		} else {
			t.Log("\tnot expected: crmservice.SetPolicy() was called unexpected number of times.")
		}
	})
}
