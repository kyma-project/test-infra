// package main contains code for Gardener GCP SA secret rotation
package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/kyma-project/test-infra/pkg/gcp/iam"
	"github.com/kyma-project/test-infra/pkg/gcp/pubsub"
	"github.com/kyma-project/test-infra/pkg/gcp/secretmanager"

	"cloud.google.com/go/compute/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	projectID             string
	secretManagerService  *secretmanager.Service
	serviceAccountService *iam.Service
	kubernetesClient      kubernetes.Interface
)

func main() {
	var err error
	ctx := context.Background()

	projectID, err = metadata.ProjectIDWithContext(ctx)
	if err != nil {
		panic("failed to retrieve GCP Project ID, error: " + err.Error())
	}

	secretManagerService, err = secretmanager.NewService(ctx)
	if err != nil {
		panic("failed creating Secret Manager client, error: " + err.Error())
	}

	serviceAccountService, err = iam.NewService(ctx)
	if err != nil {
		panic("failed creating IAM client, error: " + err.Error())
	}

	http.HandleFunc("/", RotateGardenerServiceAccount)
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

// RotateGardenerServiceAccount manages GCP SA rotation in Secret Manager and Kubernetes cluster
func RotateGardenerServiceAccount(w http.ResponseWriter, r *http.Request) {
	var ok bool
	var m pubsub.Message
	var secretRotateMessage pubsub.SecretRotateMessage
	var GardenerSASecretData iam.ServiceAccountJSON

	ctx := context.Background()

	dryRun := false
	keys, ok := r.URL.Query()["dry_run"]
	if ok && keys[0] == "true" {
		log.Printf("Dry run enabled")
		dryRun = true
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("io.ReadAll: %v", err)
		http.Error(w, "Couldn't read request body", http.StatusBadRequest)
		return
	}
	// byte slice unmarshalling handles base64 decoding.
	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("json.Unmarshal: %v", err)
		http.Error(w, "Couldn't unmarshal request body", http.StatusBadRequest)
		return
	}

	if m.Message.Attributes["eventType"] != "SECRET_ROTATE" {
		log.Printf("Unsupported event type: %s, quitting", m.Message.Attributes["eventType"])
		return
	}

	err = json.Unmarshal(m.Message.Data, &secretRotateMessage)
	if err != nil {
		log.Printf("failed to unmarshal message data field, error: %s", err)
		http.Error(w, "Couldn't unmarshal request message", http.StatusBadRequest)
		return
	}

	if secretRotateMessage.Labels["type"] != "gardener-service-account" {
		log.Printf("Unsupported secret type: %s, quitting", secretRotateMessage.Labels["type"])
		return
	}

	//check if rest of required labels is present
	var kubeconfigSecretName string
	if kubeconfigSecretName, ok = secretRotateMessage.Labels["kubeconfig-secret"]; !ok {
		log.Printf("Missing kubeconfig-secret label, quitting")
		http.Error(w, "Missing kubeconfig-secret label", http.StatusBadRequest)
		return
	}

	var gardenerSecretName string
	if gardenerSecretName, ok = secretRotateMessage.Labels["gardener-secret"]; !ok {
		log.Printf("Missing gardener-secret label, quitting")
		http.Error(w, "Missing gardener-secret label", http.StatusBadRequest)
		return
	}

	var gardenerSecretNamespace string
	if gardenerSecretNamespace, ok = secretRotateMessage.Labels["gardener-secret-namespace"]; !ok {
		log.Printf("Missing gardener-secret-namespace label, quitting")
		http.Error(w, "Missing gardener-secret-namespace label", http.StatusBadRequest)
		return
	}

	// get kubeconfig
	// TODO
	kubeconfigSecretPath := "projects/" + projectID + "/secrets/" + kubeconfigSecretName
	kubeconfig, err := secretManagerService.GetLatestSecretVersionData(kubeconfigSecretPath)
	if err != nil {
		log.Printf("Failed to get %s kubeconfig secret data: %s", kubeconfigSecretPath, err)
		http.Error(w, "Failed to get kubeconfig secret", http.StatusBadRequest)
		return
	}

	// allow mocking kubernetesClient with kubernetes.fake library
	if kubernetesClient == nil {
		kubernetesClient, err = getKubernetesClientFromConfig([]byte(kubeconfig))
		if err != nil {
			log.Printf("failed to create new Kubernetes client: %s", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	}

	//get latest Gardener SA secret version data
	log.Printf("Retrieving latest version of secret: %s", secretRotateMessage.Name)
	secretDataString, err := secretManagerService.GetLatestSecretVersionData(secretRotateMessage.Name)
	if err != nil {
		log.Printf("failed to retrieve latest version of a secret %s, error: %s", secretRotateMessage.Name, err.Error())
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal([]byte(secretDataString), &GardenerSASecretData)
	if err != nil {
		log.Printf("failed to unmarshal secret JSON field, error: %s", err.Error())
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// get list of all previous secret versions
	secretVersions, err := secretManagerService.ListSecretVersions(secretRotateMessage.Name)
	if err != nil {
		log.Printf("Could not get list of secret versions: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// create new key
	serviceAccountPath := "projects/" + GardenerSASecretData.ProjectID + "/serviceAccounts/" + GardenerSASecretData.ClientEmail
	var newKeyBytes []byte
	if !dryRun {
		log.Printf("creating new key for service account %s", serviceAccountPath)
		newKeyBytes, err = serviceAccountService.CreateNewServiceAccountKey(serviceAccountPath)
		if err != nil {
			log.Printf("failed to create new key for Service Account, error: %s", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	} else {
		log.Printf("[Dry run] creating new key for service account %s", serviceAccountPath)
	}

	// check for all necessary labels

	// push to SM
	if !dryRun {
		log.Printf("Adding new secret version to secret %s", secretRotateMessage.Name)
		_, err = secretManagerService.AddSecretVersion(secretRotateMessage.Name, newKeyBytes)
		if err != nil {
			log.Printf("failed to create new %s secret version, error: %s", secretRotateMessage.Name, err.Error())
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	} else {
		log.Printf("[Dry run] Adding new secret version to secret %s", secretRotateMessage.Name)
	}

	// get SA from Gardener
	kubeSecret, err := kubernetesClient.CoreV1().Secrets(gardenerSecretNamespace).Get(ctx, gardenerSecretName, metav1.GetOptions{})
	if err != nil {
		log.Printf("failed to retrieve Gardener %s secret, error: %s", gardenerSecretName, err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	kubeSecret.Data["serviceaccount.json"] = newKeyBytes

	// push updated SA to gardener]
	if !dryRun {
		log.Printf("Updating %s Gardener secret", gardenerSecretName)
		_, err = kubernetesClient.CoreV1().Secrets(gardenerSecretNamespace).Update(ctx, kubeSecret, metav1.UpdateOptions{})
		if err != nil {
			log.Printf("failed to update %s Service Account secret in Gardener, error: %s", gardenerSecretName, err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	} else {
		log.Printf("[Dry run] Updating %s Gardener secret", gardenerSecretName)
	}
	// delete old keys
	if !dryRun {
		log.Println("Destroying old keys")
		for _, secretVersion := range secretVersions.Versions {
			if secretVersion.State == "ENABLED" {
				versionDataString, err := secretManagerService.GetSecretVersionData(secretVersion.Name)
				if err != nil {
					log.Printf("Couldn't get payload of a %s secret: %s", secretVersion.Name, err)
					http.Error(w, "Bad Request", http.StatusBadRequest)
					return
				}

				var versionData iam.ServiceAccountJSON
				err = json.Unmarshal([]byte(versionDataString), &versionData)
				if err != nil {
					log.Printf("failed to unmarshal secret JSON field, error: %s", err)
					http.Error(w, "Bad Request", http.StatusBadRequest)
					return
				}

				// get client_email
				oldKeyPath := "projects/" + versionData.ProjectID + "/serviceAccounts/" + versionData.ClientEmail + "/keys/" + versionData.PrivateKeyID
				log.Printf("Looking for service account %s", oldKeyPath)

				err = serviceAccountService.DeleteKey(oldKeyPath)
				if err != nil {
					log.Printf("Could not delete %v key: %s", oldKeyPath, err)
				}
			}
		}
	}

	// destroy old SM
	if !dryRun {
		for _, secretVersion := range secretVersions.Versions {
			if secretVersion.State == "ENABLED" {
				secretManagerService.DestroySecretVersion(secretVersion.Name)
			}
		}
	} else {
		log.Printf("[Dry run] destroying old versions of %s secret", secretRotateMessage.Name)
	}
}

func getKubernetesClientFromConfig(kubeconfig []byte) (*kubernetes.Clientset, error) {
	kubernetesConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(kubernetesConfig)
}
