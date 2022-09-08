package secretmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/api/option"
	gcpsecretmanager "google.golang.org/api/secretmanager/v1"
)

func serveHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("url: %s ;Query: %s ;Body: %s; %s", r.URL.Path, r.URL.Query().Encode(), r.Body, r.Method)

}

func TestNewService(t *testing.T) {
	ctx := context.Background()
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	_, err := NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(fakeServer.URL))
	if err != nil {
		t.Errorf("Couldn't create new service: %s", err)
	}
}

func TestAddSecretVersion(t *testing.T) {
	ctx := context.Background()
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := gcpsecretmanager.ProjectsSecretsAddVersionCall{}
		b, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(b)
	}))
	service, err := NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(fakeServer.URL))
	if err != nil {
		t.Errorf("Couldn't create new service: %s", err)
	}

	_, err = service.AddSecretVersion("project/projectName/secret/secretName", []byte("secretData"))
	if err != nil {
		t.Errorf("Couldn't create new secret version: %s", err)
	}
}

func TestGetAllSecrets(t *testing.T) {
	ctx := context.Background()
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := gcpsecretmanager.ListSecretsResponse{}
		b, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "unable to marshal request: "+err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(b)
	}))
	service, err := NewService(ctx, option.WithoutAuthentication(), option.WithEndpoint(fakeServer.URL))
	if err != nil {
		t.Errorf("Couldn't create new service: %s", err)
	}

	_, err = service.GetAllSecrets("pro", "")
	if err != nil {
		t.Errorf("Couldn't get all secrets: %s", err)
	}
}
