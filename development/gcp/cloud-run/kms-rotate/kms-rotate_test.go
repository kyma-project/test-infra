package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/storage"
	"github.com/kyma-project/test-infra/development/gcp/pkg/pubsub"
	"google.golang.org/api/option"
)

type fakeKMSServer struct {
}

func (s *fakeKMSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

type fakeStorageServer struct {
}

func (s *fakeStorageServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestRotateKMSKey(t *testing.T) {

	var err error
	ctx := context.Background()

	tests := []struct {
		name string
		// todo
		requestBody pubsub.Message
	}{
		{},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kmsServerStruct := fakeKMSServer{}
			kmsServer := httptest.NewServer(&kmsServerStruct)
			kmsService, err = kms.NewKeyManagementClient(ctx, option.WithEndpoint(kmsServer.URL))
			if err != nil {
				t.Errorf("Couldn't set up fake kms service: %s", err)
			}

			storageServerStruct := fakeStorageServer{}
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

		})
	}

}
