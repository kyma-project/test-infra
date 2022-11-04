// package main contains code for KMS GCP symmetric key secret re-encrypting
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "cloud.google.com/go/kms/apiv1/kmspb"
	"cloud.google.com/go/storage"
	"github.com/go-playground/validator/v10"
	"google.golang.org/api/iterator"
)

// Config contains function configuration provided through POST JSON
type Config struct {
	Project      string `json:"project" validate:"required,min=1"`
	Location     string `json:"location" validate:"required,min=1"`
	BucketName   string `json:"bucketName" validate:"required,min=1"`
	BucketPrefix string `json:"bucketPrefix,omitempty"`
	Keyring      string `json:"keyring" validate:"required,min=1"`
	Key          string `json:"key" validate:"required,min=1"`
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
		log.Fatal("failed creating KMS client, error: " + err.Error())
	}
	defer kmsService.Close()

	storageService, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatal("failed creating KMS client, error: " + err.Error())
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

// RotateKMSKey function manages GCP KMS rotation with bucket files re-signing
func RotateKMSKey(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// get data from POST JSON
	body, err := io.ReadAll(r.Body)
	if err != nil {
		showError(w, http.StatusBadRequest, "Couldn't read request body: %v", err)
		return
	}

	var conf Config

	json.Unmarshal(body, &conf)

	validate := validator.New()
	err = validate.Struct(conf)
	if err != nil {
		showError(w, http.StatusBadRequest, "Missing values in config: %v", err)
		return
	}

	keyPath := "projects/" + conf.Project + "/locations/" + conf.Location + "/keyRings/" + conf.Keyring + "/cryptoKeys/" + conf.Key
	cryptoKeyRequest := &kmspb.GetCryptoKeyRequest{Name: keyPath}
	cryptoKey, err := kmsService.GetCryptoKey(ctx, cryptoKeyRequest)
	if err != nil {
		showError(w, http.StatusInternalServerError, "Couldn't get %s crypto key: %v", keyPath, err)
		return
	}
	primaryVersion := cryptoKey.GetPrimary()

	keyIteratorRequest := &kmspb.ListCryptoKeyVersionsRequest{Parent: keyPath}
	keyIterator := kmsService.ListCryptoKeyVersions(ctx, keyIteratorRequest)
	keyVersions, err := getKeyVersions(keyIterator)
	if err != nil {
		showError(w, http.StatusInternalServerError, "Couldn't iterate over %s key versions: %v", keyPath, err)
		return
	}
	if len(keyVersions) < 2 {
		log.Printf("Less than two enabled key versions, quitting")
		return
	}

	bucket := storageService.Bucket(conf.BucketName)

	err = rotateFiles(ctx, bucket, conf.BucketPrefix, keyPath)
	if err != nil {
		showError(w, http.StatusInternalServerError, "Couldn't rotate files in the %s bucket: %v", conf.BucketName, err)
		return
	}

	err = destroyOldKeyVersions(ctx, primaryVersion, keyVersions)
	if err != nil {
		showError(w, http.StatusInternalServerError, "Couldn't delete old %s key versions: %v", keyPath, err)
		return
	}
}

func showError(w http.ResponseWriter, statusCode int, format string, args ...interface{}) {
	errorMessage := fmt.Sprintf(format, args...)
	log.Println(errorMessage)
	http.Error(w, errorMessage, statusCode)
}

func rotateFiles(ctx context.Context, bucket *storage.BucketHandle, bucketPrefix, keyPath string) error {
	// for all files in bucket dir
	query := &storage.Query{}
	if bucketPrefix != "" {
		query.Prefix = bucketPrefix
	}

	bucketIterator := bucket.Objects(ctx, query)

	for {
		attrs, err := bucketIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("iterator error: %v", err)
		}

		o := bucket.Object(attrs.Name)
		encryptedData, err := getBucketObjectData(ctx, o)
		if err != nil {
			return fmt.Errorf("couldn't get %s object data: %v", attrs.Name, err)
		}

		encryptResponse, err := reencrypt(ctx, keyPath, encryptedData)
		if err != nil {
			return fmt.Errorf("couldn't re-encrypt %s object data: %v", attrs.Name, err)
		}

		err = setBucketObjectData(ctx, o, encryptResponse.Ciphertext)
		if err != nil {
			return fmt.Errorf("couldn't update %s object data: %v", attrs.Name, err)
		}
	}
	return nil
}

func getKeyVersions(keyIterator *kms.CryptoKeyVersionIterator) ([]*kmspb.CryptoKeyVersion, error) {
	var keyVersions []*kmspb.CryptoKeyVersion
	for nextVer, err := keyIterator.Next(); err != iterator.Done; nextVer, err = keyIterator.Next() {
		if err != nil && err != iterator.Done {
			return nil, err
		}
		if nextVer.State == kmspb.CryptoKeyVersion_ENABLED {
			keyVersions = append(keyVersions, nextVer)
		}
	}
	return keyVersions, nil
}

// getBucketObjectData reads data from a bucket object
func getBucketObjectData(ctx context.Context, o *storage.ObjectHandle) ([]byte, error) {
	reader, err := o.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	encryptedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	reader.Close()

	return encryptedData, nil
}

// setBucketObjectData writes data to a bucket object
func setBucketObjectData(ctx context.Context, o *storage.ObjectHandle, data []byte) error {
	writer := o.NewWriter(ctx)
	_, err := writer.Write(data)
	writer.Close()
	return err
}

// reencrypt takes in encrypted data and return the same data encrypted with the primary version of a key
func reencrypt(ctx context.Context, keyPath string, encryptedData []byte) (*kmspb.EncryptResponse, error) {
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
	return encryptResponse, nil
}

// destroyOldKeyVersions destroys all versions of key except the primary version
func destroyOldKeyVersions(ctx context.Context, primaryVersion *kmspb.CryptoKeyVersion, keyVersions []*kmspb.CryptoKeyVersion) error {
	for _, keyVersion := range keyVersions {
		if keyVersion.Name != primaryVersion.Name {
			destructionRequest := &kmspb.DestroyCryptoKeyVersionRequest{Name: keyVersion.Name}
			_, err := kmsService.DestroyCryptoKeyVersion(ctx, destructionRequest)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
