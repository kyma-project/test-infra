package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "cloud.google.com/go/kms/apiv1/kmspb"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	storageraw "google.golang.org/api/storage/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	testProjectName = "test-project"
	testLocation    = "europe-west3"
	testBucketName  = "test-bucket"
	testKeyringName = "test-keyring"
	testKeyName     = "test-key"
	testPrefix      = "certificates/"
)

type fakeKeyVersion struct {
	//	name  string
	state kmspb.CryptoKeyVersion_CryptoKeyVersionState
}

type fakeFile struct {
	// name       string
	keyVersion string
}

type bucketUpload struct {
	Bucket string `yaml:"bucket,omitempty"`
	Name   string `yaml:"name,omitempty"`
}

func getSortedKeyVersionNames(keyVersions map[string]*fakeKeyVersion) []string {
	var keyVersionNames []string
	for versionName := range keyVersions {
		keyVersionNames = append(keyVersionNames, versionName)
	}
	sort.Strings(keyVersionNames)
	return keyVersionNames
}

type fakeKMSServer struct {
	kmspb.UnimplementedKeyManagementServiceServer
	keyVersions map[string]*fakeKeyVersion
}

func (f *fakeKMSServer) GetCryptoKey(ctx context.Context, req *kmspb.GetCryptoKeyRequest) (*kmspb.CryptoKey, error) {
	// fmt.Printf("GetCryptoKey %s\n", req.Name)
	resp := &kmspb.CryptoKey{}
	if len(f.keyVersions) > 0 {
		keyVersionNames := getSortedKeyVersionNames(f.keyVersions)
		latestKeyVersionName := keyVersionNames[len(keyVersionNames)-1]
		primaryVersionName := req.Name + "/cryptoKeyVersions/" + latestKeyVersionName
		resp = &kmspb.CryptoKey{Primary: &kmspb.CryptoKeyVersion{Name: primaryVersionName}}
	}
	return resp, nil
}

func (f *fakeKMSServer) ListCryptoKeyVersions(ctx context.Context, req *kmspb.ListCryptoKeyVersionsRequest) (*kmspb.ListCryptoKeyVersionsResponse, error) {
	fmt.Printf("ListCryptoKeyVersions %s\n", req.Parent)
	resp := &kmspb.ListCryptoKeyVersionsResponse{}
	if len(f.keyVersions) > 0 {
		keyVersionNames := getSortedKeyVersionNames(f.keyVersions)
		for _, keyVersionName := range keyVersionNames {
			currentKeyVersion := f.keyVersions[keyVersionName]
			currentKeyVersionName := req.Parent + "/cryptoKeyVersions/" + keyVersionName
			cryptoKeyVersion := &kmspb.CryptoKeyVersion{Name: currentKeyVersionName, State: currentKeyVersion.state}
			resp.CryptoKeyVersions = append(resp.CryptoKeyVersions, cryptoKeyVersion)
		}
	}
	return resp, nil
}

func (f *fakeKMSServer) Decrypt(ctx context.Context, req *kmspb.DecryptRequest) (*kmspb.DecryptResponse, error) {
	resp := &kmspb.DecryptResponse{}
	resp.Plaintext = []byte("decrypted")
	return resp, nil
}

func (f *fakeKMSServer) Encrypt(ctx context.Context, req *kmspb.EncryptRequest) (*kmspb.EncryptResponse, error) {
	resp := &kmspb.EncryptResponse{}
	fmt.Printf("KMS encrypt  %s, %s\n", req.Name, string(req.Plaintext))
	keyVersionNames := getSortedKeyVersionNames(f.keyVersions)
	latestKey := keyVersionNames[len(keyVersionNames)-1]
	resp.Ciphertext = []byte(latestKey)
	return resp, nil
}

func (f *fakeKMSServer) DestroyCryptoKeyVersion(ctx context.Context, req *kmspb.DestroyCryptoKeyVersionRequest) (*kmspb.CryptoKeyVersion, error) {
	nameList := strings.Split(req.Name, "/")
	keyVersionName := nameList[9]

	resp := &kmspb.CryptoKeyVersion{Name: req.Name}
	f.keyVersions[keyVersionName].state = kmspb.CryptoKeyVersion_DESTROY_SCHEDULED
	resp.State = kmspb.CryptoKeyVersion_DESTROY_SCHEDULED
	return resp, nil
}

type fakeStorageServer struct {
	prefix                   string
	files                    map[string]*fakeFile
	unknownEndpointCallCount int
}

func (s *fakeStorageServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	objectsPathRegex := regexp.MustCompile("^/b/(.*)/o$")

	filePathRegex := regexp.MustCompile("^/" + testBucketName + "/(.*)$")
	uploadPathRegex := regexp.MustCompile("^/upload/storage/v1/b/" + testBucketName + "/o$")

	switch path := r.URL.Path; {
	case objectsPathRegex.MatchString(path):
		{
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to parse response body: "+path, 500)
			}

			objects := storageraw.Objects{Kind: "storage#objects"}

			for fileName := range s.files {
				if s.prefix != "" && !strings.HasPrefix(fileName, s.prefix) {
					continue
				}
				object := &storageraw.Object{
					Kind:   "storage#object",
					Bucket: testBucketName,
					Name:   fileName,
				}
				objects.Items = append(objects.Items, object)
			}

			fmt.Printf("BUCKET url: %s ;Query: %s ;Body: %s; %s\n", r.URL.Path, r.URL.Query().Encode(), string(body), r.Method)

			b, err := json.Marshal(objects)
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)
		}
	case filePathRegex.MatchString(path):
		{
			// simple file serving
			nameList := strings.SplitN(path, "/", 3)
			filename := nameList[2]

			w.Write([]byte(s.files[filename].keyVersion))

		}
	case uploadPathRegex.MatchString(path):
		{
			fmt.Printf("UPLOAD url: %s ;Query: %s ; %s\n", r.URL.Path, r.URL.Query().Encode(), r.Method)

			contentType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || !strings.HasPrefix(contentType, "multipart/") {
				http.Error(w, "expecting a multipart message", http.StatusBadRequest)
				return
			}

			multipartReader := multipart.NewReader(r.Body, params["boundary"])
			defer r.Body.Close()

			var newData string
			var parsedJSONdata bucketUpload

			for {
				part, err := multipartReader.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					http.Error(w, "couldn't read multipart part", http.StatusBadRequest)
					return
				}
				defer part.Close()

				partData, err := io.ReadAll(part)
				if err != nil {
					http.Error(w, "failed to read content of the part", http.StatusBadRequest)
					return
				}

				switch part.Header.Get("Content-Type") {
				case "application/json":
					err = json.Unmarshal(partData, &parsedJSONdata)
					if err != nil {
						fmt.Printf("Cannot unmarshal upload JSON %s", err)
						http.Error(w, "cannot unmarshal upload JSON", http.StatusBadRequest)
						return
					}
				case "text/plain; charset=utf-8":
					newData = string(partData)
				default:
					fmt.Printf("unknown content type %s", part.Header.Get("Content-Type"))
					http.Error(w, "unknown part content type", http.StatusBadRequest)
					return
				}
			}
			s.files[parsedJSONdata.Name].keyVersion = newData

		}
	default:
		{
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to parse response body: "+path, 500)
			}
			fmt.Printf("unknown url: %s ;Query: %s ;Body: %s; %s\n", r.URL.Path, r.URL.Query().Encode(), string(body), r.Method)
			http.Error(w, "unknown URL: "+path, 500)
			s.unknownEndpointCallCount++
		}
	}
}

// TestRotateKMSKey tests RotateKMSKey function
func TestRotateKMSKey(t *testing.T) {

	//var err error
	ctx := context.Background()

	defaultRequestBody := Config{
		Project:    testProjectName,
		Location:   testLocation,
		BucketName: testBucketName,
		Keyring:    testKeyringName,
		Key:        testKeyName,
	}

	defaultPrefixedRequestBody := Config{
		Project:      testProjectName,
		Location:     testLocation,
		BucketName:   testBucketName,
		Keyring:      testKeyringName,
		Key:          testKeyName,
		BucketPrefix: testPrefix,
	}

	tests := []struct {
		name                string
		requestBody         Config
		keyVersions         map[string]*fakeKeyVersion
		expectedKeyVersions map[string]*fakeKeyVersion
		files               map[string]*fakeFile
	}{
		{
			name:                "One enabled key, no files",
			requestBody:         defaultRequestBody,
			keyVersions:         map[string]*fakeKeyVersion{"1": {kmspb.CryptoKeyVersion_ENABLED}},
			expectedKeyVersions: map[string]*fakeKeyVersion{"1": {kmspb.CryptoKeyVersion_ENABLED}},
		},
		{
			name:                "Two enabled keys, no files",
			requestBody:         defaultRequestBody,
			keyVersions:         map[string]*fakeKeyVersion{"1": {kmspb.CryptoKeyVersion_ENABLED}, "2": {kmspb.CryptoKeyVersion_ENABLED}},
			expectedKeyVersions: map[string]*fakeKeyVersion{"1": {kmspb.CryptoKeyVersion_DESTROY_SCHEDULED}, "2": {kmspb.CryptoKeyVersion_ENABLED}},
		},
		{
			name:                "Two enabled keys, one file",
			requestBody:         defaultRequestBody,
			keyVersions:         map[string]*fakeKeyVersion{"1": {kmspb.CryptoKeyVersion_ENABLED}, "2": {kmspb.CryptoKeyVersion_ENABLED}},
			expectedKeyVersions: map[string]*fakeKeyVersion{"1": {kmspb.CryptoKeyVersion_DESTROY_SCHEDULED}, "2": {kmspb.CryptoKeyVersion_ENABLED}},
			files:               map[string]*fakeFile{"cert": {"1"}},
		},
		{
			name:                "Two enabled keys, one file, query with prefix",
			requestBody:         defaultPrefixedRequestBody,
			keyVersions:         map[string]*fakeKeyVersion{"1": {kmspb.CryptoKeyVersion_ENABLED}, "2": {kmspb.CryptoKeyVersion_ENABLED}},
			expectedKeyVersions: map[string]*fakeKeyVersion{"1": {kmspb.CryptoKeyVersion_DESTROY_SCHEDULED}, "2": {kmspb.CryptoKeyVersion_ENABLED}},
			files:               map[string]*fakeFile{"notcert": {"data"}, "certificates/cert": {"1"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kmsServerStruct := &fakeKMSServer{keyVersions: test.keyVersions}
			l, err := net.Listen("tcp", "localhost:0")
			if err != nil {
				t.Fatal(err)
			}
			gsrv := grpc.NewServer()
			kmspb.RegisterKeyManagementServiceServer(gsrv, kmsServerStruct)
			kmsEndpointURL := l.Addr().String()
			go func() {
				if err := gsrv.Serve(l); err != nil {
					t.Errorf("couldn't set up fake KMS server %s", err)
				}
			}()
			kmsService, err = kms.NewKeyManagementClient(ctx, option.WithEndpoint(kmsEndpointURL), option.WithoutAuthentication(), option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
			if err != nil {
				t.Errorf("Couldn't set up fake kms service: %s", err)
			}

			storageServerStruct := fakeStorageServer{files: test.files, prefix: test.requestBody.BucketPrefix}
			storageServer := httptest.NewServer(&storageServerStruct)
			storageService, err = storage.NewClient(ctx, option.WithoutAuthentication(), option.WithEndpoint(storageServer.URL))
			if err != nil {
				t.Errorf("Couldn't set up fake storage service: %s", err)
			}

			rr := httptest.NewRecorder()

			pubsubMessageByte, err := json.Marshal(test.requestBody)
			if err != nil {
				t.Errorf("Couldn't marshall request message: %s", err)
			}

			req := httptest.NewRequest("GET", "/", strings.NewReader(string(pubsubMessageByte)))
			req.Header.Add("Content-Type", "application/json")

			RotateKMSKey(rr, req)

			if rr.Code != 200 {
				t.Errorf("Error HTTP response: %d: %s", rr.Code, rr.Body.String())
			}

			if storageServerStruct.unknownEndpointCallCount > 0 {
				t.Errorf("Unhandled storage calls: %d", storageServerStruct.unknownEndpointCallCount)
			}

			if !reflect.DeepEqual(test.keyVersions, test.expectedKeyVersions) {
				t.Errorf("List of key versions differs, %v, want %v", test.keyVersions, test.expectedKeyVersions)
			}

			keyNames := getSortedKeyVersionNames(test.keyVersions)
			latestKeyVersion := keyNames[len(keyNames)-1]
			for filePath, f := range test.files {
				if strings.HasPrefix(filePath, test.requestBody.BucketPrefix) {
					if f.keyVersion != latestKeyVersion {
						t.Errorf("Incorrect version of key used to sign %s file: %s", filePath, f.keyVersion)
					}
				} else {
					if f.keyVersion == latestKeyVersion {
						t.Errorf("Incorrect file was signed: %s", filePath)
					}
				}
			}

		})
	}
}
