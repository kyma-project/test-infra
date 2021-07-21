package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

type ExternalSecretData struct {
	Key      string `json:"key,omitempty"`
	Name     string `json:"name,omitempty"`
	Version  string `json:"version,omitempty"`
	Property string `json:"property,omitempty"`
	IsBinary bool   `json:"isBinary,omitempty"`
}

type ExternalSecretSpec struct {
	BackendType string
	Data        []ExternalSecretData
	ProjectId   string `json:"projectId"`
}

type ExternalSecret struct {
	ApiVersion string
	Kind       string
	Metadata   v1.ObjectMeta
	Spec       ExternalSecretSpec
	Status     v1.Status
}

type ExternalSecretsList struct {
	ApiVersion string
	Items      []ExternalSecret
	Kind       string
	Metadata   v1.ListMeta
}
type options struct {
	kubeconfig string
}

func gatherOptions() options {
	o := options{}
	flag.StringVar(&o.kubeconfig, "kubeconfig", "", "Path to kubeconfig file.")
	flag.Parse()
	return o
}

func main() {
	o := gatherOptions()

	config, err := clientcmd.BuildConfigFromFlags("", o.kubeconfig)
	exitOnError(err, "while loading kubeconfig")

	client, err := kubernetes.NewForConfig(config)
	exitOnError(err, "while creating kube client")

	namespaces, err := client.CoreV1().Namespaces().List(context.Background(), v1.ListOptions{})
	exitOnError(err, "while reading namespaces list")

	externalSecretsSuccesful := true
	secretsDeclaredAsExternal := true
	exitCode := 0

	for _, namespace := range namespaces.Items {
		// get all externalSecrets in a namespace
		externalSecretsJSON, err := client.RESTClient().Get().AbsPath("/apis/kubernetes-client.io/v1").Namespace(namespace.Name).Resource("externalsecrets").DoRaw(context.Background())
		exitOnError(err, "while reading externalsecrets list")

		var externalSecretsList ExternalSecretsList
		err = json.Unmarshal(externalSecretsJSON, &externalSecretsList)
		exitOnError(err, "while unmarshalling externalSecrets list")

		// generate a list of names for easier comparison with rest of the secrets
		var externalSecretsNames []string
		for _, externalSecret := range externalSecretsList.Items {
			if externalSecret.Status.Status != "SUCCESS" {
				// externalSecretName := externalSecret.Metadata.Name
				fmt.Printf("ExternalSecret \"%s\" in %s namespace failed to synchronize with status \"%s\"\n", externalSecret.Metadata.Name, namespace.Name, externalSecret.Status.Status)
				externalSecretsSuccesful = false
			}

			externalSecretsNames = append(externalSecretsNames, externalSecret.Metadata.Name)
		}

		// at this point we have nice JSON response which we can parse further
		// .items | length == 40
		// jq '.items[0].status.status' halamix2_box_of_wonders/sekrety/ugliest_thing.json

		secrets, err := client.CoreV1().Secrets(namespace.Name).List(context.Background(), v1.ListOptions{})
		exitOnError(err, "while reading namespaces list")
		for _, sec := range secrets.Items {
			// get only user-created secrets
			if sec.Type == "Opaque" {
				// fmt.Printf("%s:\t%s, %s\n", namespace.Name, sec.Name, sec.Type)
				if !nameInSlice(sec.Name, externalSecretsNames) {
					fmt.Printf("Secret \"%s\" in %s namespace was not declared as ExternalSecret\n", sec.Name, namespace.Name)
					secretsDeclaredAsExternal = false
				}
			}
		}

		//break
	}
	if !externalSecretsSuccesful {
		fmt.Println("At least one externalsecret was not synchronized succesfully")
		exitCode += 1
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

func nameInSlice(name string, slice []string) bool {
	for _, fragment := range slice {
		if name == fragment {
			return true
		}
	}
	return false
}
