// package main contains code for KMS GCP symmetric key secret re-encrypting
package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type Config struct {
	Project  string `yaml:"project"`
	Location string `yaml:"location"`
	Keyring  string `yaml:"keyring"`
	Key      string `yaml:"key"`
}

var (
	kmsService     *kms.KeyManagementClient
	storageService *storage.Client
)

func main() {
	var err error
	ctx := context.Background()

	kmsService, err = kms.NewKeyManagementClient(ctx)
	if err != nil {
		panic("failed creating KMS client, error: " + err.Error())
	}
	defer kmsService.Close()

	storageService, err = storage.NewClient(ctx)
	if err != nil {
		panic("failed creating KMS client, error: " + err.Error())
	}

	http.HandleFunc("/", RotateKMSKey)
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	// Start HTTP server.
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func RotateKMSKey(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// get data from POST JSON
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("io.ReadAll: %v", err)
		http.Error(w, "Couldn't read request body", http.StatusBadRequest)
		return
	}

	var conf Config

	json.Unmarshal(body, &conf)

	// projects/*/locations/*/keyRings/*/cryptoKeys/*

	keyPath := "projects/" + conf.Project + "/locations/" + conf.Location + "/keyRings/" + conf.Keyring + "/cryptoKeys/" + conf.Key
	req := &kmspb.ListCryptoKeyVersionsRequest{Parent: keyPath}
	keyIterator := kmsService.ListCryptoKeyVersions(ctx, req)
	var keyVersions []*kmspb.CryptoKeyVersion
	enabledVersions := 0
	for nextVer, err := keyIterator.Next(); err != iterator.Done; nextVer, err = keyIterator.Next() {
		if err != nil && err != iterator.Done {
			log.Printf("key version iterator: %v", err)
			http.Error(w, "Couldn't iterate over key versions", http.StatusBadRequest)
			return
		}
		keyVersions = append(keyVersions, nextVer)
		nextVer.Prima
		if nextVer.State == kmspb.CryptoKeyVersion_ENABLED {
			enabledVersions++
		}
	}
	if enabledVersions <= 1 {
		log.Printf("Nothing to do, quitting")
		return
	}

}
