package serviceaccount

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount/automock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iam/v1"
)

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
			mockIAM := &automock.IAM{}
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
			mockIAM := &automock.IAM{}
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
			mockIAM := &automock.IAM{}
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
		safqdn            string
		resource          string
		prefix            string
		project           string
		keyname           string
		key               *iam.ServiceAccountKey
		serviceaccountkey *iam.ServiceAccountKey
	}
	tv := testvalue{
		safqdn:   "test-sa@test-project.iam.gserviceaccount.com",
		resource: "projects/-/serviceAccounts/test-sa@test-project.iam.gserviceaccount.com",
		prefix:   "",
		project:  "test-project",
		keyname:  "test-key",
		key: &iam.ServiceAccountKey{
			KeyAlgorithm: "KEY_ALG_RSA_2048",
			KeyOrigin:    "GOOGLE_PROVIDED",
			KeyType:      "USER_MANAGED",
			Name:         "projects/sap-kyma-neighbors-dev/serviceAccounts/credstest-dekiel-sa@sap-kyma-neighbors-dev.iam.gserviceaccount.com/keys/46e2a51e15685d1ee6ce6f203bceaeb42fed4fe6", PrivateKeyData: "ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsCiAgInByb2plY3RfaWQiOiAic2FwLWt5bWEtbmVpZ2hib3JzLWRldiIsCiAgInByaXZhdGVfa2V5X2lkIjogIjQ2ZTJhNTFlMTU2ODVkMWVlNmNlNmYyMDNiY2VhZWI0MmZlZDRmZTYiLAogICJwcml2YXRlX2tleSI6ICItLS0tLUJFR0lOIFBSSVZBVEUgS0VZLS0tLS1cbk1JSUV2UUlCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktjd2dnU2pBZ0VBQW9JQkFRQ09lcG9qcHkyaUw1WjFcbnNiakFsM3BSS3NqWU80TmFaY1NCOHpJa05oRFZadGpybFRZQ1Nla1BTRVVMS0dGK2xsTzVueTYzVUtRelVmTWtcbk1yZEZGaE52NmtpOUM5aDRaMDZTSDFnQ0dJdW95UWNTRDg2eERNbHRIV01yQTQ2Q09hZ01ucUQxZ2RwbzdveU9cbngwYytCQWp2d2lyd1lOanhpQ0h5dnE5UTZQUUlFeDBIOFk2cktLUGh3T05EekpZbHpwV1o2ckY1K2F1clAzY0RcbllQejhWbVdNNXF1eWdPR1hOY0ZCWmM4MjBtNHVjeTJRSXI4TXI4YUJCeHlIWnlZZW12QkhXYkFqRFFuS1FBaGNcbnVVZ0NQOHZGODY1V2RpN1l5T0FYdkx1bG91WHBkckNrMVFVUUE2dk91ZEE1RmIrNTd2c2l4UmVDOGt4SnVMbXhcbjJxYVNuZFU5QWdNQkFBRUNnZ0VBQlUwZWFHTkxYaCs4RHZsTlQxbVFUa2U2R2R1T04xdkpzcUt6Z0tmcWgrMWNcbnc0dWIza2x1bjhISytkeGFoa3prc2JSVy9XUFNTUVVtSGhJR3p2Um14ZGRrSEZ0a1V4U0s4RFdiUHducTNNUE1cbjNNeFVSVExkY3NKWGFVMm5obEdZa2ZUVm5RZ1hTNDVORWduc0JwL1dHK0xXV1hjQU9kUUM2Vm9oZFVuNW90OXJcbnc0QU5FK1dDVDJxSjc3VWFmNkJaSEdHWVB5OG9yZ28ya1g3M00zY25qemVSUnFYbWFHT3ZmOS9GT1ZNSks5UjlcbndNeUlIMSthWWVvUGRlSldOWU9JRVFGaFErd0RaQ0oxQnJiTlJGUmlGRHF2YlA5RUM5OXRRbTAxeC9HSU5yY21cbjc3WEtRTFBYeDhpZ3RrYjlaQ2pVSlBMT3VTS0w2MXRWREViQ3JlV3hNUUtCZ1FER2FZd3BvZ05pRTJRNGxJOWRcbi9tMjJMcTkyZDd5SFEycDZwbS8wRDA5U3dqYjVBSjlJblFzdUJKL0EyVXVTbjN0OCtBWllySk4xSjdrTDhMcm1cbjdtV1dyRGkvRC9GZEI2QU1qUmlxR3ZsVlJ0YnlEWklnUzYvazJHekUyRW1STjU2dzZ4QVVLTTZ5MEdSeGk1MC9cblQzNkR2SDZ0ZVhmakFuWEwzRjZhQzJMTjBRS0JnUUMzMVJZWXhtc3J3YlRIdCt3azd1aFFtdVlyUDF3L2FiejVcbmxIcnF0L0VIZk9VaXo4aU5oNXpvOE05VGs3cTRkelZMNzVTZTgzNzRwcE1PNDIxdUVGNFdEMmVtNkhITzVYRjhcbkdhT0JnY3F3c3VKTGt6SFY0RmtZODg5L3E1eFkyalFXQ2ZRd0pEZ2hGRURnbEhhUlphcHBqZTdPczlHdDZXcFpcbnJiY285UkdQclFLQmdRQ0hHc1A0Ylh2RVF0UVJ1d2RNeDcxSk9zejc3RmlSK3BQODVHeURVaEYvbHdQNzFqS2dcbkxWKzVmQ2lVRnVMZytud0tBcEcvdS9QRTZNR1dvZHVDK0g1d2RPRkhLTUgveVB0dzBIc2xDYTBTSm1TaStoNndcbm94a295VDUzWTVma3JHMEFwMitSYXFBbEhzWG1rMTBHQ1VscXh1V3psbXpPUlpTVXRvQXNnT2hNb1FLQmdBNFBcbkNjV1RSeGJ0bFhuQW94cWYrcnhQWEZMcVlZK244bi9UenlLc05vNndDb1lEQmY5czQ1OGM2MzRreWg3Wlh3WVRcbnFIWVBnU0phK3R3a29IWE9ZcU9sUWZRTnlzWmIzYlh6OEFFemYrRExqV3JpTXVsOFl0UDVzV0MrS3hMUWZUTkNcblI1NTI1cVFBL0lVd1ZYRUJLV3N4STVaRFFrSGVtL2VIeFg0b1g5TnhBb0dBSUZwK1FMUUJzNy96NnpadldFV3RcbnRjNUFpOGtNTzZxNnloTUY2N1RYbFBTSHNIcWljUHlacWl5T3hmVFBGdmZVQkQ5cHM1cSs3SnpxcTAyRTEzbExcbmFWb28xSkNodUY5MWdJVC9KOEt3czlNVDFMTWpzSWFKdnBPM2RHNUZUcjZaZEd0K3IvaEE1VjJQVHFtc0h5Q3pcbkpITk1WY2V6NU9SaURqY3VLWnR4eFJnPVxuLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLVxuIiwKICAiY2xpZW50X2VtYWlsIjogImNyZWRzdGVzdC1kZWtpZWwtc2FAc2FwLWt5bWEtbmVpZ2hib3JzLWRldi5pYW0uZ3NlcnZpY2VhY2NvdW50LmNvbSIsCiAgImNsaWVudF9pZCI6ICIxMDczMTI4NDUxNjkzNDc0NDQ0MDYiLAogICJhdXRoX3VyaSI6ICJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20vby9vYXV0aDIvYXV0aCIsCiAgInRva2VuX3VyaSI6ICJodHRwczovL29hdXRoMi5nb29nbGVhcGlzLmNvbS90b2tlbiIsCiAgImF1dGhfcHJvdmlkZXJfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9vYXV0aDIvdjEvY2VydHMiLAogICJjbGllbnRfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9yb2JvdC92MS9tZXRhZGF0YS94NTA5L2NyZWRzdGVzdC1kZWtpZWwtc2ElNDBzYXAta3ltYS1uZWlnaGJvcnMtZGV2LmlhbS5nc2VydmljZWFjY291bnQuY29tIgp9Cg==",
			PrivateKeyType:  "TYPE_GOOGLE_CREDENTIALS_FILE",
			PublicKeyData:   "",
			ValidAfterTime:  "2020-02-19T22:10:45Z",
			ValidBeforeTime: "2030-02-16T22:10:45Z",
			ServerResponse:  googleapi.ServerResponse{},
			ForceSendFields: make([]string, 0),
			NullFields:      make([]string, 0)},
		serviceaccountkey: &iam.ServiceAccountKey{
			KeyAlgorithm: "KEY_ALG_RSA_2048",
			KeyOrigin:    "GOOGLE_PROVIDED",
			KeyType:      "USER_MANAGED",
			Name:         "projects/sap-kyma-neighbors-dev/serviceAccounts/credstest-dekiel-sa@sap-kyma-neighbors-dev.iam.gserviceaccount.com/keys/46e2a51e15685d1ee6ce6f203bceaeb42fed4fe6", PrivateKeyData: "ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsCiAgInByb2plY3RfaWQiOiAic2FwLWt5bWEtbmVpZ2hib3JzLWRldiIsCiAgInByaXZhdGVfa2V5X2lkIjogIjQ2ZTJhNTFlMTU2ODVkMWVlNmNlNmYyMDNiY2VhZWI0MmZlZDRmZTYiLAogICJwcml2YXRlX2tleSI6ICItLS0tLUJFR0lOIFBSSVZBVEUgS0VZLS0tLS1cbk1JSUV2UUlCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktjd2dnU2pBZ0VBQW9JQkFRQ09lcG9qcHkyaUw1WjFcbnNiakFsM3BSS3NqWU80TmFaY1NCOHpJa05oRFZadGpybFRZQ1Nla1BTRVVMS0dGK2xsTzVueTYzVUtRelVmTWtcbk1yZEZGaE52NmtpOUM5aDRaMDZTSDFnQ0dJdW95UWNTRDg2eERNbHRIV01yQTQ2Q09hZ01ucUQxZ2RwbzdveU9cbngwYytCQWp2d2lyd1lOanhpQ0h5dnE5UTZQUUlFeDBIOFk2cktLUGh3T05EekpZbHpwV1o2ckY1K2F1clAzY0RcbllQejhWbVdNNXF1eWdPR1hOY0ZCWmM4MjBtNHVjeTJRSXI4TXI4YUJCeHlIWnlZZW12QkhXYkFqRFFuS1FBaGNcbnVVZ0NQOHZGODY1V2RpN1l5T0FYdkx1bG91WHBkckNrMVFVUUE2dk91ZEE1RmIrNTd2c2l4UmVDOGt4SnVMbXhcbjJxYVNuZFU5QWdNQkFBRUNnZ0VBQlUwZWFHTkxYaCs4RHZsTlQxbVFUa2U2R2R1T04xdkpzcUt6Z0tmcWgrMWNcbnc0dWIza2x1bjhISytkeGFoa3prc2JSVy9XUFNTUVVtSGhJR3p2Um14ZGRrSEZ0a1V4U0s4RFdiUHducTNNUE1cbjNNeFVSVExkY3NKWGFVMm5obEdZa2ZUVm5RZ1hTNDVORWduc0JwL1dHK0xXV1hjQU9kUUM2Vm9oZFVuNW90OXJcbnc0QU5FK1dDVDJxSjc3VWFmNkJaSEdHWVB5OG9yZ28ya1g3M00zY25qemVSUnFYbWFHT3ZmOS9GT1ZNSks5UjlcbndNeUlIMSthWWVvUGRlSldOWU9JRVFGaFErd0RaQ0oxQnJiTlJGUmlGRHF2YlA5RUM5OXRRbTAxeC9HSU5yY21cbjc3WEtRTFBYeDhpZ3RrYjlaQ2pVSlBMT3VTS0w2MXRWREViQ3JlV3hNUUtCZ1FER2FZd3BvZ05pRTJRNGxJOWRcbi9tMjJMcTkyZDd5SFEycDZwbS8wRDA5U3dqYjVBSjlJblFzdUJKL0EyVXVTbjN0OCtBWllySk4xSjdrTDhMcm1cbjdtV1dyRGkvRC9GZEI2QU1qUmlxR3ZsVlJ0YnlEWklnUzYvazJHekUyRW1STjU2dzZ4QVVLTTZ5MEdSeGk1MC9cblQzNkR2SDZ0ZVhmakFuWEwzRjZhQzJMTjBRS0JnUUMzMVJZWXhtc3J3YlRIdCt3azd1aFFtdVlyUDF3L2FiejVcbmxIcnF0L0VIZk9VaXo4aU5oNXpvOE05VGs3cTRkelZMNzVTZTgzNzRwcE1PNDIxdUVGNFdEMmVtNkhITzVYRjhcbkdhT0JnY3F3c3VKTGt6SFY0RmtZODg5L3E1eFkyalFXQ2ZRd0pEZ2hGRURnbEhhUlphcHBqZTdPczlHdDZXcFpcbnJiY285UkdQclFLQmdRQ0hHc1A0Ylh2RVF0UVJ1d2RNeDcxSk9zejc3RmlSK3BQODVHeURVaEYvbHdQNzFqS2dcbkxWKzVmQ2lVRnVMZytud0tBcEcvdS9QRTZNR1dvZHVDK0g1d2RPRkhLTUgveVB0dzBIc2xDYTBTSm1TaStoNndcbm94a295VDUzWTVma3JHMEFwMitSYXFBbEhzWG1rMTBHQ1VscXh1V3psbXpPUlpTVXRvQXNnT2hNb1FLQmdBNFBcbkNjV1RSeGJ0bFhuQW94cWYrcnhQWEZMcVlZK244bi9UenlLc05vNndDb1lEQmY5czQ1OGM2MzRreWg3Wlh3WVRcbnFIWVBnU0phK3R3a29IWE9ZcU9sUWZRTnlzWmIzYlh6OEFFemYrRExqV3JpTXVsOFl0UDVzV0MrS3hMUWZUTkNcblI1NTI1cVFBL0lVd1ZYRUJLV3N4STVaRFFrSGVtL2VIeFg0b1g5TnhBb0dBSUZwK1FMUUJzNy96NnpadldFV3RcbnRjNUFpOGtNTzZxNnloTUY2N1RYbFBTSHNIcWljUHlacWl5T3hmVFBGdmZVQkQ5cHM1cSs3SnpxcTAyRTEzbExcbmFWb28xSkNodUY5MWdJVC9KOEt3czlNVDFMTWpzSWFKdnBPM2RHNUZUcjZaZEd0K3IvaEE1VjJQVHFtc0h5Q3pcbkpITk1WY2V6NU9SaURqY3VLWnR4eFJnPVxuLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLVxuIiwKICAiY2xpZW50X2VtYWlsIjogImNyZWRzdGVzdC1kZWtpZWwtc2FAc2FwLWt5bWEtbmVpZ2hib3JzLWRldi5pYW0uZ3NlcnZpY2VhY2NvdW50LmNvbSIsCiAgImNsaWVudF9pZCI6ICIxMDczMTI4NDUxNjkzNDc0NDQ0MDYiLAogICJhdXRoX3VyaSI6ICJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20vby9vYXV0aDIvYXV0aCIsCiAgInRva2VuX3VyaSI6ICJodHRwczovL29hdXRoMi5nb29nbGVhcGlzLmNvbS90b2tlbiIsCiAgImF1dGhfcHJvdmlkZXJfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9vYXV0aDIvdjEvY2VydHMiLAogICJjbGllbnRfeDUwOV9jZXJ0X3VybCI6ICJodHRwczovL3d3dy5nb29nbGVhcGlzLmNvbS9yb2JvdC92MS9tZXRhZGF0YS94NTA5L2NyZWRzdGVzdC1kZWtpZWwtc2ElNDBzYXAta3ltYS1uZWlnaGJvcnMtZGV2LmlhbS5nc2VydmljZWFjY291bnQuY29tIgp9Cg==",
			PrivateKeyType:  "TYPE_GOOGLE_CREDENTIALS_FILE",
			PublicKeyData:   "",
			ValidAfterTime:  "2020-02-19T22:10:45Z",
			ValidBeforeTime: "2030-02-16T22:10:45Z",
			ServerResponse:  googleapi.ServerResponse{},
			ForceSendFields: make([]string, 0),
			NullFields:      make([]string, 0)},
	}
	t.Run("CreateSAKey should create key without errors.", func(t *testing.T) {
		mockIAM := &automock.IAM{}
		mockIAM.On("CreateSAKey", tv.resource, &iam.CreateServiceAccountKeyRequest{}).Return(tv.serviceaccountkey, nil)
		defer mockIAM.AssertExpectations(t)
		client := NewClient(tv.prefix, mockIAM)
		key, err := client.CreateSAKey(tv.safqdn)
		if test := assert.Nilf(t, err, "\tnot expected: CreateSAKey() returned not nil error."); test {
			t.Log("\texpected: CreateSAKey() returned nil error.")
		}
		if test := assert.Equal(t, tv.key, key, "\not expected: CreateSAKey() returned not expected key value."); test {
			t.Log("\texpected: CreateSAKey() returned expected key value.")
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
	t.Run("CreateSAKey should fail because got error from iamservice.CreateSAKey.", func(t *testing.T) {
		mockIAM := &automock.IAM{}
		mockIAM.On("CreateSAKey", tv.resource, &iam.CreateServiceAccountKeyRequest{}).Return(nil, errors.New("CreateSAKey failed GCP test error"))
		defer mockIAM.AssertExpectations(t)
		client := NewClient(tv.prefix, mockIAM)
		key, err := client.CreateSAKey(tv.safqdn)
		if test := assert.Empty(t, key, "\tnot expected: CreateSAKey() returned not empty key."); test {
			t.Log("\texpected: CreateSAKey() returned empty key.")
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
