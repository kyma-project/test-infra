package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kyma-project/test-infra/development/gcp/pkg/secretmanager"
	"github.com/kyma-project/test-infra/development/gcp/pkg/secretversionsmanager"
	"google.golang.org/api/option"
)

type ServiceAccountJSON struct {
	Type             string `json:"type"`
	ProjectID        string `json:"project_id"`
	PrivatekayID     string `json:"private_key_id"`
	PrivateKey       string `json:"private_key"`
	ClientEmail      string `json:"client_email"`
	ClientID         string `json:"client_id"`
	AuthURL          string `json:"auth_uri"`
	TokenURI         string `json:"token_uri"`
	AuthProviderCert string `json:"auth_provider_x509_cert_url"`
	ClientCert       string `json:"client_x509_cert_url"`
}

func main() {
	ctx :=
		context.Background()
	secretManagerService, err := secretmanager.NewService(ctx, option.WithCredentialsFile("/Users/i542853/Documents/sa/sa-secret-update/sap-kyma-prow-aee5ebc38d8d.json"))
	if err != nil {
		panic(fmt.Sprintf("failed creating Secret Manager client, error: %s", err.Error()))
	}
	secretVersionManagerService := secretversionsmanager.NewService(secretManagerService)
	secretlatestVersionPath := "projects/351981214969/secrets/piotr-test-rotation/versions/latest"
	secretDataString, err := secretVersionManagerService.GetSecretVersionData(secretlatestVersionPath)
	if err != nil {
		fmt.Printf("failed to retreive latest version of a secret %s, error: %s", secretlatestVersionPath, err.Error())
		os.Exit(1)
	}

	fmt.Printf("Unmarshalling %s secret\n", secretlatestVersionPath)
	var secretData ServiceAccountJSON
	decodedSecretDataString, err := base64.StdEncoding.DecodeString(secretDataString)
	err = json.Unmarshal([]byte(decodedSecretDataString), &secretData)
	if err != nil {
		fmt.Printf("failed to unmarshal secret JSON field, error: %s", err.Error())
		fmt.Printf("%s\n", secretDataString)
		os.Exit(1)
	}
}
