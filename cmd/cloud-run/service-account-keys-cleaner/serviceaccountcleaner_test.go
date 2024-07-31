package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/test-infra/pkg/gcp/iam"
	"github.com/kyma-project/test-infra/pkg/gcp/secretmanager"

	gcpiam "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	gcpsecretmanager "google.golang.org/api/secretmanager/v1"
)

var (
	testingProjectName = "testing"
)

type fakeSecretVersion struct {
	Data  string
	Date  string
	State string
}

func createFakeServiceAccountJSONData(email, key, state string) (string, error) {
	data := iam.ServiceAccountJSON{
		ProjectID:    projectID,
		ClientEmail:  email,
		PrivateKeyID: key,
	}
	dataString, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(dataString), nil
}

type fakeSecret struct {
	Labels   map[string]string
	Versions map[string]*fakeSecretVersion
}

type fakeSecretServer struct {
	secrets                  map[string]*fakeSecret
	unknownEndpointCallCount int
}

// naiveParseLabelFilter fakes filter parsing on server-side.
// Expects filter in labels.name=value format
// returns name, value, error
func naiveParseLabelFilter(filter string) (string, string, error) {
	first, value, equalFilter := strings.Cut(filter, "=")
	if !equalFilter {
		return "", "", errors.New("unexpected filter:" + filter + ", expected = symbol")
	}
	labels, name, twoPartFilter := strings.Cut(first, ".")
	if (labels != "labels") || !twoPartFilter {
		return "", "", errors.New("unexpected filter:" + filter + ", expected = symbol")
	}
	return name, value, nil
}

// naiveParseStateFilter fakes filter parsing on server-side.
// Expects filter in key:value format
// returns key, value, error
func naiveParseSimpleFilter(filter string) (string, string, error) {
	key, value, colonFilter := strings.Cut(filter, ":")
	if !colonFilter {
		return "", "", errors.New("unexpected filter:" + filter + ", expected = symbol")
	}
	return key, value, nil
}

func createSecretName(project, secret string) string {
	return "projects/" + project + "/secrets/" + secret
}

func createSecretVersionName(project, secret, version string) string {
	return createSecretName(project, secret) + "/versions/" + version
}

func (s *fakeSecretServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	versionsRegex := regexp.MustCompile("/v1/projects/(.*)/secrets/(.*)/versions$")
	versionAccessRegex := regexp.MustCompile("/v1/projects/(.*)/secrets/(.*)/versions/(.*):access$")
	versionDestroyRegex := regexp.MustCompile("/v1/projects/(.*)/secrets/(.*)/versions/(.*):destroy$")

	switch path := r.URL.Path; {
	case path == "/v1/projects/"+testingProjectName+"/secrets":
		{
			filterLabel := ""
			filterLabelValue := ""
			if r.URL.Query().Get("filter") != "" {
				filterLabel, filterLabelValue, err = naiveParseLabelFilter(r.URL.Query().Get("filter"))
				if err != nil {
					http.Error(w, "unable to parse labels filter: "+err.Error(), http.StatusBadRequest)
				}
			}

			resp := gcpsecretmanager.ListSecretsResponse{}
			for name, s := range s.secrets {
				if filterLabel != "" {
					if val, ok := s.Labels[filterLabel]; ok && val != filterLabelValue {
						continue
					}
				}
				resp.Secrets = append(resp.Secrets, &gcpsecretmanager.Secret{Name: createSecretName(testingProjectName, name), Labels: s.Labels})
			}
			b, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)
		}
	case versionsRegex.MatchString(path):
		{
			versionsResponse := gcpsecretmanager.ListSecretVersionsResponse{}
			// unknown url: /v1/secret_one_version/versions ;Query: filter=state%3Aenabled
			filterKey := ""
			filterValue := ""
			if r.URL.Query().Get("filter") != "" {
				filterKey, filterValue, err = naiveParseSimpleFilter(r.URL.Query().Get("filter"))
				if err != nil {
					http.Error(w, "unable to parse version state filter: "+err.Error(), http.StatusBadRequest)
				}
				if filterKey != "state" {
					http.Error(w, "Bad filter key, expected state, got "+filterKey, http.StatusBadRequest)
				}
			}
			projectName := strings.Split(path, "/")[3]
			secretName := strings.Split(path, "/")[5]
			if secret, ok := s.secrets[secretName]; ok {
				var versionKeys []string
				for versionName := range secret.Versions {
					versionKeys = append(versionKeys, versionName)
				}
				sort.Strings(versionKeys)

				for _, versionName := range versionKeys {
					currentVersion := secret.Versions[versionName]
					if filterValue != "" {
						if currentVersion.State != filterValue {
							continue
						}
					}
					// secret versions are returned in reverse, from latest to oldest
					version := &gcpsecretmanager.SecretVersion{Name: createSecretVersionName(projectName, secretName, versionName), CreateTime: currentVersion.Date, State: currentVersion.State}
					versionsResponse.Versions = append([]*gcpsecretmanager.SecretVersion{version}, versionsResponse.Versions...)
				}
			} else {
				http.Error(w, "unable to find secret: "+secretName, http.StatusBadRequest)
			}

			b, err := json.Marshal(versionsResponse)
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)
		}
	case versionAccessRegex.MatchString(path):
		{
			projectName := strings.Split(path, "/")[3]
			secretName := strings.Split(path, "/")[5]
			versionNameRaw := strings.Split(path, "/")[7]
			versionName := strings.Split(versionNameRaw, ":")[0]

			secret, ok := s.secrets[secretName]
			if !ok {
				http.Error(w, "unable to find secret: "+secretName, http.StatusBadRequest)
				return
			}
			version, ok2 := secret.Versions[versionName]
			if !ok2 {
				http.Error(w, "unable to find secret "+secretName+" version "+versionName, http.StatusBadRequest)
				return
			}
			payload := gcpsecretmanager.SecretPayload{Data: version.Data}
			versionResponse := gcpsecretmanager.AccessSecretVersionResponse{Name: createSecretVersionName(projectName, secretName, versionName), Payload: &payload}
			b, err := json.Marshal(versionResponse)
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)
		}
	case versionDestroyRegex.MatchString(path):
		{
			// projectName := strings.Split(path, "/")[3]
			secretName := strings.Split(path, "/")[5]
			versionNameRaw := strings.Split(path, "/")[7]
			versionName := strings.Split(versionNameRaw, ":")[0]

			ver := s.secrets[secretName].Versions[versionName]
			ver.State = "destroyed"

			b, err := json.Marshal(gcpsecretmanager.SecretVersion{})
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)
		}
	default:
		{
			fmt.Printf("unknown url: %s ;Query: %s ;Body: %s; %s\n", r.URL.Path, r.URL.Query().Encode(), r.Body, r.Method)
			http.Error(w, "unknown URL: "+r.URL.Path, 500)
			s.unknownEndpointCallCount++
		}
	}
}

type fakeIAMServer struct {
	keys                     map[string]map[string]bool
	unknownEndpointCallCount int
}

func (s *fakeIAMServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	keyPathRegex := regexp.MustCompile("/v1/projects/(.*)/serviceAccounts/(.*)/keys/(.*)$")
	switch path := r.URL.Path; {
	case keyPathRegex.MatchString(path):
		{
			if r.Method != "DELETE" {
				return
			}
			// projectName := strings.Split(path, "/")[3]
			serviceAccountName := strings.Split(path, "/")[5]
			keyName := strings.Split(path, "/")[7]

			if _, ok := s.keys[serviceAccountName]; !ok {
				fmt.Printf("unable to find service account %s\n", serviceAccountName)
				// don't return here, as the function can still delete secret version even if the key was deleted manually
			}

			delete(s.keys[serviceAccountName], keyName)

			b, err := json.Marshal(gcpiam.Empty{})
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)
		}
	default:
		{
			fmt.Printf("unknown url: %s ;Query: %s ;Body: %s; %s\n", r.URL.Path, r.URL.Query().Encode(), r.Body, r.Method)
			http.Error(w, "unknown URL: "+r.URL.Path, 500)
			s.unknownEndpointCallCount++
		}
	}
}

func TestServiceAccountKeysCleaner(t *testing.T) {
	var err error
	ctx := context.Background()
	projectID = testingProjectName
	fakeSecretEmail := "email"

	fakeSecretKey := "secretKeyID"
	fakeSecretVersionData, err := createFakeServiceAccountJSONData(fakeSecretEmail, fakeSecretKey, "enabled")
	if err != nil {
		t.Errorf("could not generate fake secret version data: %s", err)
	}

	fakeSecretKey2 := "secretKeyID2"
	fakeSecretVersionData2, err := createFakeServiceAccountJSONData(fakeSecretEmail, fakeSecretKey, "enabled")
	if err != nil {
		t.Errorf("could not generate fake secret version data: %s", err)
	}

	timeThreeHoursAgo := time.Now().Add(time.Duration(-3) * time.Hour).UTC().Format("2006-01-02T15:04:05.000000Z")
	timeTwoHoursAgo := time.Now().Add(time.Duration(-2) * time.Hour).UTC().Format("2006-01-02T15:04:05.000000Z")
	time59MinutesAgo := time.Now().Add(time.Duration(-59) * time.Minute).UTC().Format("2006-01-02T15:04:05.000000Z")

	tests := []struct {
		name             string
		secrets          map[string]*fakeSecret
		expectedSecrets  map[string]*fakeSecret
		keys             map[string]map[string]bool
		expectedKeys     map[string]map[string]bool
		requestBody      string
		expectedResponse string
	}{
		{
			name:             "empty secrets list",
			secrets:          make(map[string]*fakeSecret),
			expectedSecrets:  make(map[string]*fakeSecret),
			keys:             make(map[string]map[string]bool),
			expectedKeys:     make(map[string]map[string]bool),
			requestBody:      "",
			expectedResponse: "",
		},
		{
			name: "secret without labels",
			secrets: map[string]*fakeSecret{"secret_no_label": {
				Labels:   map[string]string{},
				Versions: map[string]*fakeSecretVersion{"1": {Data: fakeSecretVersionData, Date: timeThreeHoursAgo, State: "enabled"}},
			}},
			expectedSecrets: map[string]*fakeSecret{"secret_no_label": {
				Labels:   map[string]string{},
				Versions: map[string]*fakeSecretVersion{"1": {Data: fakeSecretVersionData, Date: timeThreeHoursAgo, State: "enabled"}},
			}},
			keys:             make(map[string]map[string]bool),
			expectedKeys:     make(map[string]map[string]bool),
			requestBody:      "",
			expectedResponse: "",
		},
		{
			name: "secret with correct labels, one enabled version",
			secrets: map[string]*fakeSecret{"secret_one_version": {
				Labels:   map[string]string{"type": "service-account"},
				Versions: map[string]*fakeSecretVersion{"1": {Data: fakeSecretVersionData, Date: timeThreeHoursAgo, State: "enabled"}},
			}},
			expectedSecrets: map[string]*fakeSecret{"secret_one_version": {
				Labels:   map[string]string{"type": "service-account"},
				Versions: map[string]*fakeSecretVersion{"1": {Data: fakeSecretVersionData, Date: timeThreeHoursAgo, State: "enabled"}},
			}},
			keys:             make(map[string]map[string]bool),
			expectedKeys:     make(map[string]map[string]bool),
			requestBody:      "",
			expectedResponse: "",
		},
		{
			name: "secret with correct labels, two enabled version, latest is not outdated",
			secrets: map[string]*fakeSecret{"secret_new": {
				Labels: map[string]string{"type": "service-account"},
				Versions: map[string]*fakeSecretVersion{
					"1": {Data: fakeSecretVersionData, Date: timeThreeHoursAgo, State: "enabled"},
					"2": {Data: fakeSecretVersionData, Date: time59MinutesAgo, State: "enabled"},
				},
			}},
			expectedSecrets: map[string]*fakeSecret{"secret_new": {
				Labels: map[string]string{"type": "service-account"},
				Versions: map[string]*fakeSecretVersion{
					"1": {Data: fakeSecretVersionData, Date: timeThreeHoursAgo, State: "enabled"},
					"2": {Data: fakeSecretVersionData, Date: time59MinutesAgo, State: "enabled"},
				},
			}},
			keys:             map[string]map[string]bool{fakeSecretEmail: {fakeSecretKey: true, fakeSecretKey2: true}},
			expectedKeys:     map[string]map[string]bool{fakeSecretEmail: {fakeSecretKey: true, fakeSecretKey2: true}},
			requestBody:      "",
			expectedResponse: "",
		},
		{
			name: "secret with correct labels, two enabled version, latest is outdated",
			secrets: map[string]*fakeSecret{"secret_outdated": {
				Labels: map[string]string{"type": "service-account"},
				Versions: map[string]*fakeSecretVersion{
					"1": {Data: fakeSecretVersionData, Date: timeThreeHoursAgo, State: "enabled"},
					"2": {Data: fakeSecretVersionData2, Date: timeTwoHoursAgo, State: "enabled"},
				},
			}},
			expectedSecrets: map[string]*fakeSecret{"secret_outdated": {
				Labels: map[string]string{"type": "service-account"},
				Versions: map[string]*fakeSecretVersion{
					"1": {Data: fakeSecretVersionData, Date: timeThreeHoursAgo, State: "destroyed"},
					"2": {Data: fakeSecretVersionData2, Date: timeTwoHoursAgo, State: "enabled"},
				},
			}},
			keys:             map[string]map[string]bool{fakeSecretEmail: {fakeSecretKey: true, fakeSecretKey2: true}},
			expectedKeys:     map[string]map[string]bool{fakeSecretEmail: {fakeSecretKey2: true}},
			requestBody:      "",
			expectedResponse: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			secretServerStruct := fakeSecretServer{secrets: test.secrets}

			secretServer := httptest.NewServer(&secretServerStruct)
			secretManagerService, err = secretmanager.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(secretServer.URL))
			if err != nil {
				t.Errorf("Couldn't set up fake secretmanager: %s", err)
			}

			iamServerStruct := fakeIAMServer{keys: test.keys}
			iamServer := httptest.NewServer(&iamServerStruct)
			serviceAccountService, err = gcpiam.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(iamServer.URL))
			if err != nil {
				t.Errorf("Couldn't set up fake serviceaccountmanager: %s", err)
			}

			req := httptest.NewRequest("GET", "/", strings.NewReader(test.requestBody))
			req.Header.Add("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			serviceAccountKeysCleaner(rr, req)

			if got := rr.Body.String(); got != test.expectedResponse {
				t.Errorf("ServiceAccountCleaner(%q) = %q, want %q", test.requestBody, got, test.expectedResponse)
			}

			if !reflect.DeepEqual(test.keys, test.expectedKeys) {
				t.Errorf("List of expected keys differs, %v, want %v", test.keys, test.expectedKeys)
			}

			if !reflect.DeepEqual(test.secrets, test.expectedSecrets) {
				t.Errorf("List of expected secrets differs, %v, want %v", test.secrets, test.expectedSecrets)
			}

			if secretServerStruct.unknownEndpointCallCount > 0 {
				t.Errorf("At least one secret manager endpoint was not faked properly")
			}

			if iamServerStruct.unknownEndpointCallCount > 0 {
				t.Errorf("At least one IAM endpoint was not faked properly")
			}
		})
	}
}
