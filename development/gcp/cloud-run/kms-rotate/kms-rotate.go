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
	Project      string `yaml:"project"`
	Location     string `yaml:"location"`
	BucketName   string `yaml:"bucketName"`
	BucketPrefix string `yaml:"bucketPrefix"`
	Keyring      string `yaml:"keyring"`
	Key          string `yaml:"key"`
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

	if conf.Project == "" {
		log.Print("project value is missing")
		http.Error(w, "project value is not provided", http.StatusBadRequest)
		return
	}
	if conf.Location == "" {
		log.Print("location value is missing")
		http.Error(w, "location value is not provided", http.StatusBadRequest)
		return
	}
	if conf.Keyring == "" {
		log.Print("keyring value is missing")
		http.Error(w, "keyring value is not provided", http.StatusBadRequest)
		return
	}
	if conf.Key == "" {
		log.Print("key value is missing")
		http.Error(w, "key value is not provided", http.StatusBadRequest)
		return
	}
	if conf.BucketName == "" {
		log.Print("bucket value is missing")
		http.Error(w, "bucket value is not provided", http.StatusBadRequest)
		return
	}

	keyPath := "projects/" + conf.Project + "/locations/" + conf.Location + "/keyRings/" + conf.Keyring + "/cryptoKeys/" + conf.Key
	cryptoKeyRequest := &kmspb.GetCryptoKeyRequest{Name: keyPath}
	cryptoKey, err := kmsService.GetCryptoKey(ctx, cryptoKeyRequest)
	if err != nil {
		log.Printf("get crypto key: %v", err)
		http.Error(w, "Couldn't get crypto key", http.StatusBadRequest)
		return
	}
	primaryVersion := cryptoKey.GetPrimary()

	keyIteratorRequest := &kmspb.ListCryptoKeyVersionsRequest{Parent: keyPath}
	keyIterator := kmsService.ListCryptoKeyVersions(ctx, keyIteratorRequest)
	var keyVersions []*kmspb.CryptoKeyVersion
	enabledVersions := 0
	for nextVer, err := keyIterator.Next(); err != iterator.Done; nextVer, err = keyIterator.Next() {
		if err != nil && err != iterator.Done {
			log.Printf("key version iterator: %v", err)
			http.Error(w, "Couldn't iterate over key versions", http.StatusBadRequest)
			return
		}
		keyVersions = append(keyVersions, nextVer)
		if nextVer.State == kmspb.CryptoKeyVersion_ENABLED {
			enabledVersions++
		}
	}
	if enabledVersions < 2 {
		log.Printf("Less than two enabled key versions, quitting")
		return
	}

	bucket := storageService.Bucket(conf.BucketName)

	// for all files in bucket dir
	query := &storage.Query{}
	if conf.BucketPrefix != "" {
		query.Prefix = conf.BucketPrefix
	}

	bucketIterator := bucket.Objects(ctx, query)

	for {
		attrs, err := bucketIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// decrypt
		// encrypt with latest
		reader, err := bucket.Object(attrs.Name).NewReader(ctx)
		if err != nil {
			log.Fatal(err)
		}
		var encryptedData []byte
		_, err = reader.Read(encryptedData)
		if err != nil {
			log.Fatal(err)
		}
		reader.Close()

		decryptRequest := &kmspb.DecryptRequest{Name: keyPath, Ciphertext: encryptedData}
		decryptResponse, err := kmsService.Decrypt(ctx, decryptRequest)
		if err != nil {
			log.Fatal(err)
		}

		encryptRequest := &kmspb.EncryptRequest{Name: keyPath, Plaintext: decryptResponse.Plaintext}
		encryptResponse, err := kmsService.Encrypt(ctx, encryptRequest)
		if err != nil {
			log.Fatal(err)
		}

		writer := bucket.Object(attrs.Name).NewWriter(ctx)
		_, err = writer.Write(encryptResponse.Ciphertext)
		if err != nil {
			log.Fatal(err)
		}
		writer.Close()
	}

	// destroy old keys
	for _, keyVersion := range keyVersions {
		if keyVersion.Name != primaryVersion.Name {
			destructionRequest := &kmspb.DestroyCryptoKeyVersionRequest{Name: keyVersion.Name}
			_, err := kmsService.DestroyCryptoKeyVersion(ctx, destructionRequest)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
