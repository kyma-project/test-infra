package serviceaccount

import (
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iam/v1"
	"testing"
)

const checkMark = "\u2713"
const ballotX = "\u2717"

var testvalues = []struct {
	prefix  string
	saname  string
	project string
}{
	{"test_prefix", "test_sa", "test_project"},
	{"", "test_sa", "test_project"},
	{"very_long_test_prefix", "very_long_sa_name", "test_project"},
}

func Test_NewClient(t *testing.T) {
	for _, tv := range testvalues {
		t.Logf("Testing with values:\n\tprefix: %s", tv.prefix)
		prefix := tv.prefix
		t.Run("NewClient() should be able to create Client object without errors.", func(t *testing.T) {
			mockIAM := &mocks.IAM{}
			client := NewClient(prefix, mockIAM)
			if test := assert.Equal(t, prefix, client.prefix, "Prefix field should be equal to passed prefix string as argument. %s", ballotX); test {
				t.Log("\tprefix field is equal to prefix string passed as argument", checkMark)
			}
			if test := assert.Equal(t, mockIAM, client.iamservice, "Imaservice field should be equal to passed argument of IAM type. %s", ballotX); test {
				t.Log("\tiamservice field is equal to iamservice IAM type object passed as argument.", checkMark)
			}
			if test := assert.NotNil(t, client.iamservice, "Iamservice field should not have nil value. %s", ballotX); test {
				t.Log("\timaservice field is not nil.", checkMark)
			}
		})
	}
}

func TestClient_CreateSA(t *testing.T) {
	for _, tv := range testvalues {
		t.Logf("Testing with values:\n\tprefix: %s\n\tsaname: %s\n\tproject: %s", tv.prefix, tv.saname, tv.project)
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
			if test := assert.Nil(t, err, "Client.CreateSA() method returned not nil error. %s", ballotX); test {
				t.Log("CreateSA() returned nil error", checkMark)
			}
			if test := assert.NotEmpty(t, sa, "Client.CrateSA() method returned empty serviceaccount object. %s", ballotX); test {
				t.Log("CrateSA() returned not empty serviceaccount object.", checkMark)
			}
			if test := mockIAM.AssertCalled(t, "CreateSA", &iam.CreateServiceAccountRequest{
				AccountId: prefixedsa,
			}, prefixedproject); test {
				t.Log("CreateSA() was called with expected arguments.", checkMark)
			} else {
				t.Errorf("CreateSA() was not called or called with unexpected arguments. %s", ballotX)
			}
			if test := mockIAM.AssertNumberOfCalls(t, "CreateSA", 1); test {
				t.Log("CreateSA() was called expected number of times.", checkMark)
			} else {
				t.Errorf("CreateSA() was called unexpected number of times. %s", ballotX)
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
			if test := assert.NotNil(t, err, "\tClient.CreateSA() method returned nil error. %s"); test {
				t.Log("\tCreateSA() returned not nil error")
			}
			if test := assert.Empty(t, sa, "\tClient.CrateSA() method returned non empty serviceaccount object."); test {
				t.Log("\tCrateSA() returned empty serviceaccount object.")
			}
			if test := mockIAM.AssertCalled(t, "CreateSA", &iam.CreateServiceAccountRequest{
				AccountId: prefixedsa,
			}, prefixedproject); test {
				t.Log("\tCreateSA() was called with expected arguments.")
			} else {
				t.Log("\tCreateSA() was not called or called with unexpected arguments.")
			}
			if test := mockIAM.AssertNumberOfCalls(t, "CreateSA", 1); test {
				t.Log("\tCreateSA() was called expected number of times.")
			} else {
				t.Log("\tCreateSA() was called unexpected number of times.")
			}
		})
	}
}
