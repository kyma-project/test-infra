package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

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

	for _, namespace := range namespaces.Items {
		// secrets, err := client.CoreV1().Secrets(namespace.Name).List(context.Background(), v1.ListOptions{})
		// exitOnError(err, "while reading namespaces list")
		// for _, sec := range secrets.Items {
		//     if sec.Type == "Opaque" {
		//         fmt.Printf("%s:\t%s, %s\n", namespace.Name, sec.Name, sec.Type)
		//     }
		// }

		// get all externalSecrets in a namespace
		data, err := client.RESTClient().Get().AbsPath("/apis/kubernetes-client.io/v1").Namespace(namespace.Name).Resource("externalsecrets").DoRaw(context.Background())
		exitOnError(err, "while reading externalsecrets list")
		// at this point we have nice JSON response which we can parse further
		// .items | length == 40
		// jq '.items[0].status.status' halamix2_box_of_wonders/sekrety/ugliest_thing.json
		fmt.Printf("%s", data)
		break
	}
}
func exitOnError(err error, context string) {
	if err != nil {
		logrus.Fatal(errors.Wrap(err, context))
	}
}
