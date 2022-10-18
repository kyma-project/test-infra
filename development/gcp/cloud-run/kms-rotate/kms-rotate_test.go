package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	storageraw "google.golang.org/api/storage/v1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	testProjectName = "test-project"
	testLocation    = "europe-west3"
	testBucketName  = "test-bucket"
	testKeyringName = "test-keyring"
	testKeyName     = "test-key"
)

type fakeKeyVersion struct {
	//	name  string
	state kmspb.CryptoKeyVersion_CryptoKeyVersionState
}

type fakeFile struct {
	// name       string
	keyVersion string
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
	fmt.Printf("GetCryptoKey %s\n", req.Name)
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
			cryptoKeyVersion := &kmspb.CryptoKeyVersion{State: currentKeyVersion.state}
			resp.CryptoKeyVersions = append(resp.CryptoKeyVersions, cryptoKeyVersion)
		}
	}

	return resp, nil
}

func (f *fakeKMSServer) Decrypt(ctx context.Context, req *kmspb.DecryptRequest) (*kmspb.DecryptResponse, error) {
	resp := &kmspb.DecryptResponse{}
	return resp, nil
}

func (f *fakeKMSServer) Encrypt(ctx context.Context, req *kmspb.EncryptRequest) (*kmspb.EncryptResponse, error) {
	resp := &kmspb.EncryptResponse{}
	return resp, nil
}

func (f *fakeKMSServer) DestroyCryptoKeyVersion(ctx context.Context, req *kmspb.DestroyCryptoKeyVersionRequest) (*kmspb.CryptoKeyVersion, error) {
	resp := &kmspb.CryptoKeyVersion{}
	return resp, nil
}

type fakeStorageServer struct {
	files                    map[string]*fakeFile
	unknownEndpointCallCount int
}

func (s *fakeStorageServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	objectsPathRegex := regexp.MustCompile("^/b/(.*)/o$")

	objectPathRegex := regexp.MustCompile("^/" + testBucketName + "/(.*)$")
	uploadPathRegex := regexp.MustCompile("^/upload/storage/v1/b/" + testBucketName + "/o$")

	switch path := r.URL.Path; {
	case objectsPathRegex.MatchString(path):
		{
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to parse response body: "+path, 500)
			}

			var objects storageraw.Objects

			object := &storageraw.Object{Name: "aa"}
			objects.Items = append(objects.Items, object)

			fmt.Printf("BUCKET url: %s ;Query: %s ;Body: %s; %s\n", r.URL.Path, r.URL.Query().Encode(), string(body), r.Method)

			b, err := json.Marshal(objects)
			if err != nil {
				http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(b)
		}
	case objectPathRegex.MatchString(path):
		{
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to parse response body: "+path, 500)
			}
			fmt.Printf("OBJECT url: %s ;Query: %s ;Body: %s; %s\n", r.URL.Path, r.URL.Query().Encode(), string(body), r.Method)
		}
	case uploadPathRegex.MatchString(path):
		{

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to parse response body: "+path, 500)
			}
			fmt.Printf("UPLOAD url: %s ;Query: %s ;Body: %s; %s\n", r.URL.Path, r.URL.Query().Encode(), string(body), r.Method)
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

	tests := []struct {
		name                string
		requestBody         Config
		keyVersions         map[string]*fakeKeyVersion
		expectedKeyVersions map[string]*fakeKeyVersion
		files               map[string]*fakeFile
	}{
		{
			name:                "test 1",
			requestBody:         defaultRequestBody,
			keyVersions:         map[string]*fakeKeyVersion{"1": {kmspb.CryptoKeyVersion_ENABLED}, "2": {kmspb.CryptoKeyVersion_ENABLED}},
			expectedKeyVersions: map[string]*fakeKeyVersion{"1": {kmspb.CryptoKeyVersion_DISABLED}, "2": {kmspb.CryptoKeyVersion_ENABLED}},
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
					panic(err)
				}
			}()
			kmsService, err = kms.NewKeyManagementClient(ctx, option.WithEndpoint(kmsEndpointURL), option.WithoutAuthentication(), option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
			if err != nil {
				t.Errorf("Couldn't set up fake kms service: %s", err)
			}

			storageServerStruct := fakeStorageServer{files: test.files}
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

			for filePath, f := range test.files {
				if f.keyVersion != "latest" {
					t.Errorf("Incorrect version of key used to sign %s file: %s", filePath, f.keyVersion)
				}
			}

		})
	}
}
