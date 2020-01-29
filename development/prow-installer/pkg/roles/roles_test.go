package roles

import (
	"errors"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/roles/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/googleapi"
	"net/http"
	"regexp"
	"testing"
)

const checkMark = "\u2713"
const ballotX = "\u2717"

var testvalues = []struct {
	saname  string
	project string
	role    string
	policy  cloudresourcemanager.Policy
}{
	{"test_sa_01", "test_project_01", "test_role_01", cloudresourcemanager.Policy{}},
	{"test_sa_02", "test_project_02", "test_role_02", cloudresourcemanager.Policy{
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
	}},
	{"test_sa_03", "test_project_03", "organizations/test_role_03", cloudresourcemanager.Policy{
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
	}},
	{"test_sa_04", "test_project_04", "roles/test_role_04", cloudresourcemanager.Policy{}},
}

func Test_New(t *testing.T) {
	t.Run("New() should create client object without errors.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		crmclient, err := New(mockCRM)
		if test := assert.IsTypef(t, &client{}, crmclient, "\tReturned client object is not a pointer to client type. %s", ballotX); test {
			t.Log("\tNew() returned pointer to object of client type.", checkMark)
		}
		if test := assert.Nilf(t, err, "\tNew() returned not nil error. %s", ballotX); test {
			t.Log("\tNew() returned nil error.", checkMark)
		}
		if test := assert.Equalf(t, mockCRM, crmclient.crmservice, "\tNew() returned client object with crmservice field other than provided as argument. %s", ballotX); test {
			t.Log("\tNew() returned client object with crmservice field same as provided as argument", checkMark)
		}
		if test := assert.Emptyf(t, crmclient.policies, "\tNew() returned client with not empty policies field. %s", ballotX); test {
			t.Log("\tNew() returned client with empty policies field", checkMark)
		}
	})
}

// NOT READY
// tests: without errors, each string arg empty, add sa to existing role, add sa to non existing role, add sa to role where it's already present, add sa to multiple roles at one function call, add sa to different roles in separate calls.
func TestClient_AddSAtoRole(t *testing.T) {
	t.Run("AddSAtoRole should add serviceaccount to role without errors.", func(t *testing.T) {

	})
	testvalues := []struct {
		saname  string
		role    string
		project string
	}{
		{"", "test_role", "test_project"},
		{"test_sa", "", "test_project"},
		{"test_sa", "test_role", ""},
	}
	for _, tv := range testvalues {
		t.Logf("Testing with values:\n\tsaname: %s\n\tproject: %s\n\trole: %s", tv.saname, tv.project, tv.role)
		t.Run("AddSAtoRole should fail because of empty mandatory argument passed.", func(t *testing.T) {

		})
	}
}

// NOT READY changed method sygnature.
func TestClient_getPolicy(t *testing.T) {
	t.Run("getPolicy() should get policy without errors.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		mockCRM.On("GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}).Return(&testvalues[1].policy, nil)
		defer mockCRM.AssertExpectations(t)
		err := client.getPolicy("test-project")
		require.Nilf(t, err, "\tgetPolicy() returned not nil error. %s", ballotX)
		t.Log("\tgetPolicy() returned nil error.", checkMark)
		require.IsTypef(t, &cloudresourcemanager.Policy{}, client.policies["test-project"], "\tgetPolicy() added policy object is not type of cloudresourcemanager.Policy. %s", ballotX)
		t.Log("\tgetPolicy() added policy of type cloudresourcemanager.Policy.", checkMark)
		//TODO: Check if change to ElementsMatch test.
		require.Equalf(t, testvalues[1].policy.Bindings, client.policies["test-project"].Bindings, "\tgetPolicy() added policy with different Bindings slice than expected. %s", ballotX)
		t.Log("\tgetPolicy() returned policy with expected Bindings slice.", checkMark)
		if test := mockCRM.AssertCalled(t, "GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\tcrmservice.GetPolicy() was called with expected arguments.", checkMark)
		} else {
			t.Log("\tcrmservice.GetPolicy() was called with unexpected arguments.", ballotX)
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\tcrmservice.GetPolicy() was called expected number of times.", checkMark)
		} else {
			t.Log("\tcrmservice.GetPolicy() was called unexpected number of times.", ballotX)
		}
	})
	t.Run("getPolicy() returned http.StatusNotModified.", func(t *testing.T) {
		notModifiedError := &googleapi.Error{
			Code:    http.StatusNotModified,
			Message: "",
			Body:    "",
			Header:  nil,
			Errors:  nil,
		}
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		client.policies["test-project"] = &testvalues[1].policy
		mockCRM.On("GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, notModifiedError)
		defer mockCRM.AssertExpectations(t)
		err := client.getPolicy("test-project")
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
		require.Truef(t, test, "\tgetPolicy() returned error did not contain http.StatusNotModified code. %s", ballotX)
		t.Log("\tgetPolicy() returned error containing http.StatusNotModified code.", checkMark)
		require.Lenf(t, client.policies, 1, "\tgetPolicy() changed number of policies hold by client object, expected only one policy.", ballotX)
		t.Log("\tgetPolicy() did not changed number od policies hold by client object.", checkMark)
		require.Equalf(t, testvalues[1].policy, *client.policies["test-project"], "\tgetPolicy() modified policies hold by client.", ballotX)
		t.Log("\tgetPolicy() did not modified policies hold by client.", checkMark)
		if test := mockCRM.AssertCalled(t, "GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\tcrmservice.GetPolicy() was called with expected arguments.", checkMark)
		} else {
			t.Log("\tcrmservice.GetPolicy() was called with unexpected arguments.", ballotX)
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\tcrmservice.GetPolicy() was called expected number of times.", checkMark)
		} else {
			t.Log("\tcrmservice.GetPolicy() was called unexpected number of times.", ballotX)
		}
	})

	t.Run("getPolicy() should fail and return not nil error.", func(t *testing.T) {
		testError := errors.New("Get test-project policy error")
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		client.policies["test-project"] = &testvalues[1].policy
		mockCRM.On("GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}).Return(nil, testError)
		defer mockCRM.AssertExpectations(t)
		err := client.getPolicy("test-project")
		require.EqualErrorf(t, err, "When downloading policy for test-project project got error: [Get test-project policy error].", "\tgetPolicy() returned error message other than expected. %s", ballotX)
		t.Log("\tgetPolicy() returned expected error message.", checkMark)
		require.Lenf(t, client.policies, 1, "\tgetPolicy() changed number of policies hold by client object, expected only one policy.", ballotX)
		t.Log("\tgetPolicy() did not changed number od policies hold by client object.", checkMark)
		require.Equalf(t, testvalues[1].policy, *client.policies["test-project"], "\tgetPolicy() modified policies hold by client.", ballotX)
		t.Log("\tgetPolicy() did not modified policies hold by client.", checkMark)
		if test := mockCRM.AssertCalled(t, "GetPolicy", "test-project", &cloudresourcemanager.GetIamPolicyRequest{}); test {
			t.Log("\tcrmservice.GetPolicy() was called with expected arguments.", checkMark)
		} else {
			t.Log("\tcrmservice.GetPolicy() was called with unexpected arguments.", ballotX)
		}
		if test := mockCRM.AssertNumberOfCalls(t, "GetPolicy", 1); test {
			t.Log("\tcrmservice.GetPolicy() was called expected number of times.", checkMark)
		} else {
			t.Log("\tcrmservice.GetPolicy() was called unexpected number of times.", ballotX)
		}
	})
}

/*
func TestClient_addToRole(t *testing.T) {
	t.Run("addToRole should add sa to role without errors.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		err := client
	})
}
*/

func TestClient_makeSafqdn(t *testing.T) {
	t.Run("makeSafqdn() should return GCP policy valid FQDN serviceaccount name.", func(t *testing.T) {
		mockCRM := &mocks.CRM{}
		client, _ := New(mockCRM)
		safqdn := client.makeSafqdn(testvalues[0].saname, testvalues[0].project)
		require.Equalf(t, fmt.Sprintf("serviceAccount:%s@%s.iam.gserviceaccount.com", testvalues[0].saname, testvalues[0].project), safqdn, "makeSafqdn() returned unexpected string value. %s", ballotX)
		t.Log("makeSafqdn() returned expected string value.", checkMark)
	})
}

func TestClient_makeRoleFullname(t *testing.T) {
	for _, tv := range testvalues {
		t.Logf("Testing with values:\n\trole: %s", tv.role)
		var role string
		if prefixed, _ := regexp.MatchString("^(roles|organizations)/.*", tv.role); prefixed {
			role = tv.role
		} else {
			role = fmt.Sprintf("roles/%s", tv.role)
		}
		t.Run("makeRoleFullname() shuold return GCP policy valid name", func(t *testing.T) {
			mockCRM := &mocks.CRM{}
			client, _ := New(mockCRM)
			rolefullname := client.makeRoleFullname(tv.role)
			require.Equalf(t, role, rolefullname, "makeRoleFullname() returned unexpected value for GCP policy role name. %s", ballotX)
			t.Log("makeRoleFulllname() returned expected GCP policy role name.", checkMark)
		})
	}
}
