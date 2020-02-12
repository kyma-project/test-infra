package serviceaccount

import (
	"errors"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iam/v1"
	"testing"
)

/*
var testvalues = []struct {
	prefix  string
	saname  string
	project string
}{
	{"test_prefix", "test_sa", "test_project"},
	{"", "test_sa", "test_project"},
	{"very_long_test_prefix", "very_long_sa_name", "test_project"},
}
*/
func Test_NewClient(t *testing.T) {
	type testvalue struct {
		prefix string
	}
	testvalues := []testvalue{
		{"test_prefix"},
		{""},
		{"very_long_test_prefix"},
	}
	for _, tv := range testvalues {
		t.Logf("\nTesting with values:\n\tprefix: %s", tv.prefix)
		prefix := tv.prefix
		t.Run("NewClient() should be able to create Client object without errors.", func(t *testing.T) {
			mockIAM := &mocks.IAM{}
			client := NewClient(prefix, mockIAM)
			if test := assert.Equal(t, prefix, client.prefix, "\tnot expected: client.prefix should be equal to passed prefix as argument."); test {
				t.Log("\texpected: client.prefix is equal to prefix passed as argument")
			}
			if test := assert.Equal(t, mockIAM, client.iamservice, "\tnot expected: client.imaservice should be equal to passed IAM argument."); test {
				t.Log("\texpected: client.iamservice is equal to iamservice IAM passed as argument.")
			}
			if test := assert.NotNil(t, client.iamservice, "\tnot expected: client.iamservice should not have nil value."); test {
				t.Log("\texpected: client.imaservice field is not nil.")
			}
		})
	}
}

func TestClient_CreateSA(t *testing.T) {
	type testvalue struct {
		prefix  string
		saname  string
		project string
	}
	testvalues := []testvalue{
		{"test_prefix", "test_sa", "test_project"},
		{"", "test_sa", "test_project"},
		{"very_long_test_prefix", "very_long_sa_name", "test_project"},
	}
	for _, tv := range testvalues {
		t.Logf("\nTesting with values:\n\tprefix: %s\n\tsaname: %s\n\tproject: %s", tv.prefix, tv.saname, tv.project)
		var prefixedsa string
		project := tv.project
		saname := tv.saname
		prefix := tv.prefix
		prefixedproject := fmt.Sprintf("projects/%s", project)
		if prefix != "" {
			prefixedsa = fmt.Sprintf("%s-%s", prefix, saname)
		} else {
			prefixedsa = saname
		}
		prefixedsa = fmt.Sprintf("%.30s", prefixedsa)
		fqdnsa := prefixedsa + "@" + project + ".iam.gserviceaccount.com"
		t.Run("CreateSA() should create serviceaccount.", func(t *testing.T) {
			mockIAM := &mocks.IAM{}
			client := NewClient(prefix, mockIAM)
			mockIAM.On("CreateSA", &iam.CreateServiceAccountRequest{
				AccountId: prefixedsa,
			}, prefixedproject).Return(&iam.ServiceAccount{Name: fqdnsa}, nil)
			defer mockIAM.AssertExpectations(t)
			sa, err := client.CreateSA(saname, project)
			if test := assert.Nil(t, err, "\tnot expected: Client.CreateSA() method returned not nil error."); test {
				t.Log("\texpected: CreateSA() returned nil error")
			}
			if test := assert.NotEmpty(t, sa, "\tnot expected: Client.CrateSA() method returned empty serviceaccount object."); test {
				t.Log("\texpected: CrateSA() returned not empty serviceaccount object.")
			}
			if test := mockIAM.AssertCalled(t, "CreateSA", &iam.CreateServiceAccountRequest{
				AccountId: prefixedsa,
			}, prefixedproject); test {
				t.Log("\texpected: CreateSA() was called with expected arguments.")
			} else {
				t.Errorf("\tnot expected: CreateSA() was not called or called with unexpected arguments.")
			}
			if test := mockIAM.AssertNumberOfCalls(t, "CreateSA", 1); test {
				t.Log("\texpected: CreateSA() was called expected number of times.")
			} else {
				t.Errorf("\tnot expected: CreateSA() was called unexpected number of times.")
			}
		})
		t.Run("CreateSA() fail and should return error", func(t *testing.T) {
			mockIAM := &mocks.IAM{}
			client := NewClient(prefix, mockIAM)
			mockIAM.On("CreateSA", &iam.CreateServiceAccountRequest{
				AccountId: prefixedsa,
			}, prefixedproject).Return(&iam.ServiceAccount{}, fmt.Errorf("Creating %s service account failed with error code from GCP.", prefixedsa))
			defer mockIAM.AssertExpectations(t)
			sa, err := client.CreateSA(saname, project)
			if test := assert.NotNil(t, err, "\tnot expected: Client.CreateSA() method returned nil error."); test {
				t.Log("\texpected: CreateSA() returned not nil error")
			}
			if test := assert.Empty(t, sa, "\tnot expected: Client.CrateSA() method returned non empty serviceaccount object."); test {
				t.Log("\texpected: CrateSA() returned empty serviceaccount object.")
			}
			if test := mockIAM.AssertCalled(t, "CreateSA", &iam.CreateServiceAccountRequest{
				AccountId: prefixedsa,
			}, prefixedproject); test {
				t.Log("\texpected: CreateSA() was called with expected arguments.")
			} else {
				t.Log("\tnot expected: CreateSA() was called with unexpected arguments.")
			}
			if test := mockIAM.AssertNumberOfCalls(t, "CreateSA", 1); test {
				t.Log("\texpected: CreateSA() was called expected number of times.")
			} else {
				t.Log("\tnot expected: CreateSA() was called unexpected number of times.")
			}
		})
	}
}

func TestClient_CreateSAKey(t *testing.T) {
	type testvalue struct {
		safqdn   string
		resource string
		prefix   string
		project  string
		keyname  string
	}
	tv := testvalue{
		safqdn:   "test-sa@test-project.iam.gserviceaccount.com",
		resource: "projects/-/serviceAccounts/test-sa@test-project.iam.gserviceaccount.com",
		prefix:   "",
		project:  "test-project",
		keyname:  "test-key",
	}
	t.Run("CreateSAKey should create key without errors.", func(t *testing.T) {
		mockIAM := &mocks.IAM{}
		mockIAM.On("CreateSAKey", tv.resource, &iam.CreateServiceAccountKeyRequest{}).Return(&iam.ServiceAccountKey{
			Name: tv.keyname,
		}, nil)
		defer mockIAM.AssertExpectations(t)
		client := NewClient(tv.prefix, mockIAM)
		key, err := client.CreateSAKey(tv.safqdn)
		if test := assert.Nilf(t, err, "\tnot expected: CreateSAKey() returned not nil error."); test {
			t.Log("\texpected: CreateSAKey() returned nil error.")
		}
		if test := assert.IsTypef(t, &iam.ServiceAccountKey{}, key, "\tnot expected: CreateSAKey() returned key not type of *iam.ServiceAccountKey"); test {
			t.Log("\texpected: CreateSAKey() returned key as expected type.")
		}
		if test := mockIAM.AssertCalled(t, "CreateSAKey", tv.resource, &iam.CreateServiceAccountKeyRequest{}); test {
			t.Log("\texpected: IAMservice.CreateSAKey() was called with expected arguments.")
		} else {
			t.Log("\tnot expected: IAMservice.CreateSAKey() was not called with expected arguments.")
		}
		if test := mockIAM.AssertNumberOfCalls(t, "CreateSAKey", 1); test {
			t.Log("\texpected: IAMservice.CreateSAKey() was called expected number of times.")
		} else {
			t.Log("\tnot expected: IAMservice.CreateSAKey() was called unexpected number of times.")
		}
		if test := assert.Equalf(t, tv.keyname, key.Name, "\tnot expected: CreateSAKey() returned key with unexpected name."); test {
			t.Log("\texpected: CreateSAKey() returned key with expected name.")
		}
	})
	t.Run("CreateSAKey should fail because got error from iamservice.CreateSAKey.", func(t *testing.T) {
		mockIAM := &mocks.IAM{}
		mockIAM.On("CreateSAKey", tv.resource, &iam.CreateServiceAccountKeyRequest{}).Return(nil, errors.New("CreateSAKey failed GCP test error"))
		defer mockIAM.AssertExpectations(t)
		client := NewClient(tv.prefix, mockIAM)
		key, err := client.CreateSAKey(tv.safqdn)
		if test := assert.Nilf(t, key, "\tnot expected: CreateSAKey() returned not nil key."); test {
			t.Log("\texpected: CreateSAKey() returned nil key.")
		}
		if test := assert.Errorf(t, err, "\tnot expected: CreateSAKey did not returned error."); test {
			t.Log("\texpected: CreateSAKey() returned error")
		}
		message1 := assert.Containsf(t, err.Error(), "CreateSAKey failed GCP test error", "\tnot expected: CreateSAKey() did not contained expected error message.")
		message2 := assert.Containsf(t, err.Error(), "When creating key for serviceaccount", "\tnot expected: CreateSAKey() did not contained expected error message.")
		if message1 && message2 {
			t.Log("\texpected: CreateSAKey() returned expected error message")
		}
		if test := mockIAM.AssertCalled(t, "CreateSAKey", tv.resource, &iam.CreateServiceAccountKeyRequest{}); test {
			t.Log("\texpected: IAMservice.CreateSAKey() was called with expected arguments.")
		} else {
			t.Log("\tnot expected: IAMservice.CreateSAKey() was not called with expected arguments.")
		}
		if test := mockIAM.AssertNumberOfCalls(t, "CreateSAKey", 1); test {
			t.Log("\texpected: IAMservice.CreateSAKey() was called expected number of times.")
		} else {
			t.Log("\tnot expected: IAMservice.CreateSAKey() was called unexpected number of times.")
		}
	})
}
