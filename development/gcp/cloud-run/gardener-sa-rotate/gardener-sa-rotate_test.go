package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	gpubsub "cloud.google.com/go/pubsub"
	"github.com/kyma-project/test-infra/development/gcp/pkg/iam"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"github.com/kyma-project/test-infra/development/gcp/pkg/secretmanager"
	gcpiam "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
	gcpsecretmanager "google.golang.org/api/secretmanager/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	testingProjectName      = "testing"
	gardenerSecretNamespace = "garden-cluster"
	saSecretLabels          = map[string]string{
		"type":                      "gardener-service-account",
		"kubeconfig-secret":         "kubeconfig-secret",
		"gardener-secret":           "sa-secret",
		"gardener-secret-namespace": gardenerSecretNamespace}
	fakeSAData0 = createFakeServiceAccountJSONData("email", "key0")
	fakeSAData  = createFakeServiceAccountJSONData("email", "key1")
	fakeSAData2 = createFakeServiceAccountJSONData("email", "key2")
)

type fakeKubernetesSecret struct {
	namespace string
	data      map[string][]byte
}

type fakeSecretVersion struct {
	Data  string
	Date  string
	State string
}

func createFakeServiceAccountJSONData(email, key string) string {
	data := iam.ServiceAccountJSON{
		ProjectID:    projectID,
		ClientEmail:  email,
		PrivateKeyID: key,
	}
	dataString, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(dataString)
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
	versionsLatestRegex := regexp.MustCompile("/v1/projects/(.*)/secrets/(.*)/versions/latest$")
	versionAccessRegex := regexp.MustCompile("/v1/projects/(.*)/secrets/(.*)/versions/(.*):access$")
	versionDestroyRegex := regexp.MustCompile("/v1/projects/(.*)/secrets/(.*)/versions/(.*):destroy$")
	versionAddRegex := regexp.MustCompile("/v1/projects/(.*)/secrets/(.*):addVersion$")

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
	case versionsLatestRegex.MatchString(path):
		{
			var secret *fakeSecret
			var ok bool

			projectName := strings.Split(path, "/")[3]
			secretName := strings.Split(path, "/")[5]
			if secret, ok = s.secrets[secretName]; !ok {
				http.Error(w, "unable to find secret"+secretName, http.StatusBadRequest)
				return
			}
			var versionKeys []string
			for versionName := range secret.Versions {
				versionKeys = append(versionKeys, versionName)
			}
			sort.Strings(versionKeys)

			versionName := versionKeys[len(versionKeys)-1]

			latestVersion := secret.Versions[versionName]

			version := &gcpsecretmanager.SecretVersion{Name: createSecretVersionName(projectName, secretName, versionName), CreateTime: latestVersion.Date, State: latestVersion.State}
			b, err := json.Marshal(version)
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

			var version *fakeSecretVersion

			secret, ok := s.secrets[secretName]
			if !ok {
				http.Error(w, "unable to find secret: "+secretName, http.StatusBadRequest)
				return
			}
			if versionName == "latest" {
				var versionKeys []string
				for versionName := range secret.Versions {
					versionKeys = append(versionKeys, versionName)
				}
				sort.Strings(versionKeys)

				if len(versionKeys) > 0 {
					versionName = versionKeys[len(versionKeys)-1]
				}
			}

			version, ok = secret.Versions[versionName]
			if !ok {
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
			ver.State = "DESTROYED"

			b, err := json.Marshal(ver)
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)
		}
	case versionAddRegex.MatchString(path):
		{
			projectName := strings.Split(path, "/")[3]
			secretNameRaw := strings.Split(path, "/")[5]
			secretName := strings.Split(secretNameRaw, ":")[0]

			data, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "unable to read request: "+err.Error(), http.StatusBadRequest)
				return
			}

			var unpackedData gcpsecretmanager.AddSecretVersionRequest
			err = json.Unmarshal(data, &unpackedData)
			if err != nil {
				http.Error(w, "unable to unmarshal request: "+err.Error(), http.StatusBadRequest)
				return
			}

			var versions []string
			for versionName := range s.secrets[secretName].Versions {
				versions = append(versions, versionName)
			}
			sort.Strings(versions)

			lastVersion, err := strconv.Atoi(versions[len(versions)-1])
			if err != nil {
				http.Error(w, "unable to convert string to int: "+err.Error(), http.StatusBadRequest)
				return
			}
			newVersionName := strconv.Itoa(lastVersion + 1)
			unpackedPayload, err := base64.StdEncoding.DecodeString(unpackedData.Payload.Data)
			if err != nil {
				http.Error(w, "unable to decode base64: "+err.Error(), http.StatusBadRequest)
				return
			}

			s.secrets[secretName].Versions[newVersionName] = &fakeSecretVersion{Data: string(unpackedPayload), Date: "", State: "ENABLED"}

			secretVersion := &gcpsecretmanager.SecretVersion{Name: createSecretVersionName(projectName, secretName, newVersionName), CreateTime: "", State: "ENABLED"}
			b, err := json.Marshal(secretVersion)
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

	keysPathRegex := regexp.MustCompile("/v1/projects/(.*)/serviceAccounts/(.*)/keys$")
	keyPathRegex := regexp.MustCompile("/v1/projects/(.*)/serviceAccounts/(.*)/keys/(.*)$")
	switch path := r.URL.Path; {
	case keysPathRegex.MatchString(path):
		{
			projectName := strings.Split(path, "/")[3]
			serviceAccountName := strings.Split(path, "/")[5]

			keyID := "key2"

			s.keys[serviceAccountName][keyID] = true

			newKey := gcpiam.ServiceAccountKey{
				Name:           "projects/" + projectName + "/serviceAccounts/" + serviceAccountName + "/keys/" + keyID,
				PrivateKeyData: base64.StdEncoding.EncodeToString([]byte(fakeSAData2)),
			}

			b, err := json.Marshal(newKey)
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)
		}
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

func TestRotateGardenerServiceAccount(t *testing.T) {
	var err error
	ctx := context.Background()
	projectID = testingProjectName

	kubeconfigSecret := &fakeSecret{Versions: map[string]*fakeSecretVersion{"1": {Data: base64.StdEncoding.EncodeToString([]byte("data")), State: "ENABLED"}}}

	tests := []struct {
		name string

		keys                      map[string]map[string]bool
		expectedKeys              map[string]map[string]bool
		secrets                   map[string]*fakeSecret
		expectedSecrets           map[string]*fakeSecret
		kuberentesSecrets         map[string]fakeKubernetesSecret
		expectedKuberentesSecrets map[string]fakeKubernetesSecret

		requestBody        pubsub.Message
		requestMessageData pubsub.SecretRotateMessage
		expectedResponse   string
	}{
		{
			name: "One old enabled SA",

			secrets: map[string]*fakeSecret{
				"sa-secret":         {Labels: saSecretLabels, Versions: map[string]*fakeSecretVersion{"1": {Data: fakeSAData, State: "ENABLED"}}},
				"kubeconfig-secret": kubeconfigSecret,
			},
			expectedSecrets: map[string]*fakeSecret{
				"sa-secret":         {Labels: saSecretLabels, Versions: map[string]*fakeSecretVersion{"1": {Data: fakeSAData, State: "DESTROYED"}, "2": {Data: fakeSAData2, State: "ENABLED"}}},
				"kubeconfig-secret": kubeconfigSecret,
			},
			keys:         map[string]map[string]bool{"email": {"key1": true}},
			expectedKeys: map[string]map[string]bool{"email": {"key2": true}},

			kuberentesSecrets:         map[string]fakeKubernetesSecret{"sa-secret": {namespace: gardenerSecretNamespace, data: map[string][]byte{"serviceaccount.json": []byte(fakeSAData)}}},
			expectedKuberentesSecrets: map[string]fakeKubernetesSecret{"sa-secret": {namespace: gardenerSecretNamespace, data: map[string][]byte{"serviceaccount.json": []byte(fakeSAData2)}}},
			requestBody:               pubsub.Message{Message: gpubsub.Message{Attributes: map[string]string{"eventType": "SECRET_ROTATE"}}},
			requestMessageData:        pubsub.SecretRotateMessage{Name: createSecretName(testingProjectName, "sa-secret"), Labels: saSecretLabels},
		},
		{
			name: "One destroyed, one old enabled SA",

			secrets: map[string]*fakeSecret{
				"sa-secret":         {Labels: saSecretLabels, Versions: map[string]*fakeSecretVersion{"1": {Data: fakeSAData0, State: "DESTROYED"}, "2": {Data: fakeSAData, State: "ENABLED"}}},
				"kubeconfig-secret": kubeconfigSecret,
			},
			expectedSecrets: map[string]*fakeSecret{
				"sa-secret":         {Labels: saSecretLabels, Versions: map[string]*fakeSecretVersion{"1": {Data: fakeSAData0, State: "DESTROYED"}, "2": {Data: fakeSAData, State: "DESTROYED"}, "3": {Data: fakeSAData2, State: "ENABLED"}}},
				"kubeconfig-secret": kubeconfigSecret,
			},
			keys:         map[string]map[string]bool{"email": {"key1": true}},
			expectedKeys: map[string]map[string]bool{"email": {"key2": true}},

			kuberentesSecrets:         map[string]fakeKubernetesSecret{"sa-secret": {namespace: gardenerSecretNamespace, data: map[string][]byte{"serviceaccount.json": []byte(fakeSAData)}}},
			expectedKuberentesSecrets: map[string]fakeKubernetesSecret{"sa-secret": {namespace: gardenerSecretNamespace, data: map[string][]byte{"serviceaccount.json": []byte(fakeSAData2)}}},
			requestBody:               pubsub.Message{Message: gpubsub.Message{Attributes: map[string]string{"eventType": "SECRET_ROTATE"}}},
			requestMessageData:        pubsub.SecretRotateMessage{Name: createSecretName(testingProjectName, "sa-secret"), Labels: saSecretLabels},
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
			serviceAccountService, err = iam.NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(iamServer.URL))
			if err != nil {
				t.Errorf("Couldn't set up fake serviceaccountmanager: %s", err)
			}

			kubernetesClient = fake.NewSimpleClientset()

			for gardenerSecretName, gardenerSecret := range test.kuberentesSecrets {
				meta := metav1.ObjectMeta{Name: gardenerSecretName, Namespace: gardenerSecret.namespace}
				secret := &v1.Secret{Type: "Opaque", Data: gardenerSecret.data, ObjectMeta: meta}
				kubernetesClient.CoreV1().Secrets(gardenerSecretNamespace).Create(ctx, secret, metav1.CreateOptions{})
			}

			rr := httptest.NewRecorder()

			secretRotateMessageByte, err := json.Marshal(test.requestMessageData)
			if err != nil {
				t.Errorf("Couldn't marshall Secret Manager message: %s", err)
			}
			test.requestBody.Message.Data = secretRotateMessageByte

			pubsubMessageByte, err := json.Marshal(test.requestBody)
			if err != nil {
				t.Errorf("Couldn't marshall pubsub message: %s", err)
			}

			req := httptest.NewRequest("GET", "/", strings.NewReader(string(pubsubMessageByte)))
			req.Header.Add("Content-Type", "application/json")

			RotateGardenerServiceAccount(rr, req)

			if got := rr.Body.String(); got != test.expectedResponse {
				t.Errorf("ServiceAccountCleaner(%v) = %q, want %q", test.requestBody, got, test.expectedResponse)
			}

			if !reflect.DeepEqual(test.keys, test.expectedKeys) {
				t.Errorf("List of expected keys differs, %v, want %v", test.keys, test.expectedKeys)
			}

			if !reflect.DeepEqual(test.secrets, test.expectedSecrets) {
				t.Errorf("List of expected secrets differs, %v, want %v", test.secrets, test.expectedSecrets)
			}

			// todo copmpare kubernetes secrets

			test.kuberentesSecrets = make(map[string]fakeKubernetesSecret)
			kubernetesSecrets, err := kubernetesClient.CoreV1().Secrets(gardenerSecretNamespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				t.Errorf("Couldn't get list of Gardener secrets: %s", err)
			}
			for _, ks := range kubernetesSecrets.Items {
				// decodedData, err := base64.RawStdEncoding.DecodeString(string(ks.Data[]))
				test.kuberentesSecrets[ks.Name] = fakeKubernetesSecret{namespace: ks.Namespace, data: ks.Data}
			}

			if !reflect.DeepEqual(test.kuberentesSecrets, test.expectedKuberentesSecrets) {
				t.Errorf("List of expected Kubernetes secrets differs, %v, want %v", test.kuberentesSecrets, test.expectedKuberentesSecrets)
			}
		})
	}
}
