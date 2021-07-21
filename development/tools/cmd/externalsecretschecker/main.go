package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

type externalSecretData struct {
	Key      string `json:"key,omitempty"`
	Name     string `json:"name,omitempty"`
	Version  string `json:"version,omitempty"`
	Property string `json:"property,omitempty"`
	IsBinary bool   `json:"isBinary,omitempty"`
}

type externalSecretSpec struct {
	BackendType string
	Data        []externalSecretData
	ProjectID   string `json:"projectId"`
}

// ExternalSecret stores one externalSecret data
type ExternalSecret struct {
	APIVersion string `json:"apiVersion"`
	Kind       string
	Metadata   v1.ObjectMeta
	Spec       externalSecretSpec
	Status     v1.Status
}

// ExternalSecretsList stores list of external secrets returned by the REST API
type ExternalSecretsList struct {
	APIVersion string `json:"apiVersion"`
	Items      []ExternalSecret
	Kind       string
	Metadata   v1.ListMeta
}
type options struct {
	skipStatusCheck  bool
	skipSecretsCheck bool
	namespaces       string
	ignoredSecrets   string
	kubeconfig       string
}

func gatherOptions() options {
	o := options{}
	flag.BoolVar(&o.skipStatusCheck, "skip-status-check", false, "Skip status check of externalSecrets.")
	// TODO this is ugly help message
	flag.BoolVar(&o.skipSecretsCheck, "skip-secrets-check", false, "Skip secret to externalSecret conenction check.")
	flag.StringVar(&o.namespaces, "namespaces", "", "names of namespaces to check, separated by comma. If empty checks all.")
	flag.StringVar(&o.ignoredSecrets, "ignored-secrets", "", "names of ignored secrets in namespace/secretName format, separated by comma.")
	flag.StringVar(&o.kubeconfig, "kubeconfig", "", "Path to kubeconfig file.")
	flag.Parse()
	return o
}

func main() {
	o := gatherOptions()
	externalSecretsSuccesful := true
	secretsDeclaredAsExternal := true
	exitCode := 0

	config, err := clientcmd.BuildConfigFromFlags("", o.kubeconfig)
	exitOnError(err, "while loading kubeconfig")

	client, err := kubernetes.NewForConfig(config)
	exitOnError(err, "while creating kube client")

	namespaces, err := client.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	exitOnError(err, "while reading namespaces list")

	var checkedNamespaces []string
	if o.namespaces != "" {
		checkedNamespaces = strings.Split(o.namespaces, ",")
	}

	ignoredSecrets := make(map[string][]string)
	ignoredSecretsSplit := strings.Split(o.ignoredSecrets, ",")
	for _, ignored := range ignoredSecretsSplit {
		split := strings.Split(ignored, "/")

		ignoredSecrets[split[0]] = append(ignoredSecrets[split[0]], split[1])
	}

	for _, namespace := range namespaces.Items {
		if len(checkedNamespaces) == 0 || nameInSlice(namespace.Name, checkedNamespaces) {
			// get all externalSecrets in a namespace
			externalSecretsJSON, err := client.RESTClient().Get().AbsPath("/apis/kubernetes-client.io/v1").Namespace(namespace.Name).Resource("externalsecrets").DoRaw(context.Background())
			exitOnError(err, "while reading externalsecrets list")

			var externalSecretsList ExternalSecretsList
			err = json.Unmarshal(externalSecretsJSON, &externalSecretsList)
			exitOnError(err, "while unmarshalling externalSecrets list")

			// generate a list of names for easier comparison with rest of the secrets
			var externalSecretsNames []string

			// check if externalSecrets synced successfully
			for _, externalSecret := range externalSecretsList.Items {
				if !o.skipStatusCheck {
					if externalSecret.Status.Status != "SUCCESS" {
						// externalSecretName := externalSecret.Metadata.Name
						fmt.Printf("ExternalSecret \"%s\" in %s namespace failed to synchronize with status \"%s\"\n", externalSecret.Metadata.Name, namespace.Name, externalSecret.Status.Status)
						externalSecretsSuccesful = false
					}
				}

				externalSecretsNames = append(externalSecretsNames, externalSecret.Metadata.Name)
			}

			// check if all Opaque secrets are also declared as externalSecrets
			if !o.skipSecretsCheck {
				secrets, err := client.CoreV1().Secrets(namespace.Name).List(context.Background(), v1.ListOptions{})
				exitOnError(err, "while reading namespaces list")
				for _, sec := range secrets.Items {
					// get only user-created secrets
					if sec.Type == "Opaque" {
						if !nameInSlice(sec.Name, ignoredSecrets[namespace.Name]) {
							if !nameInSlice(sec.Name, externalSecretsNames) {
								fmt.Printf("Secret \"%s\" in %s namespace was not declared as ExternalSecret\n", sec.Name, namespace.Name)
								secretsDeclaredAsExternal = false
							}
						}
					}
				}
			}
		}
	}

	if !externalSecretsSuccesful {
		fmt.Println("At least one externalsecret was not synchronized succesfully")
		exitCode++
	}

	if !secretsDeclaredAsExternal {
		fmt.Println("At least one secret was not declared as ExternalSecret")
		exitCode += 2
	}

	if exitCode == 0 {
		fmt.Println("No issues detected :)")
	}

	os.Exit(exitCode)
}

func exitOnError(err error, context string) {
	if err != nil {
		logrus.Fatal(errors.Wrap(err, context))
	}
}

// nameInSlice checks if the string is in slice of strings
func nameInSlice(name string, slice []string) bool {
	for _, fragment := range slice {
		if name == fragment {
			return true
		}
	}
	return false
}
