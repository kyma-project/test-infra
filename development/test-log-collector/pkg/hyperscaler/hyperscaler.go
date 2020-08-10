package hyperscaler

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Platform string

const (
	Gke             Platform = "GKE"
	Aks             Platform = "AKS"
	GardenerAzure   Platform = "gardenerAzure"
	GardenerGcp     Platform = "gardenerGcp"
	UnknownGardener Platform = "unknownGardener"
	Unknown         Platform = "unknown"
)

const (
	shootCmNamespace = "kube-config"
	shootCmName      = "shoot-info"
)

func extractHyperScalerFromCm(configmap corev1.ConfigMap) (Platform, error) {
	providerKey := "provider"
	provider, ok := configmap.Data[providerKey]
	if !ok {
		return UnknownGardener, fmt.Errorf("%s confimap in namespace %s is malformed, there's no %s key", shootCmName, shootCmNamespace, providerKey)
	}

	switch provider {
	case "azure":
		return GardenerAzure, nil
	case "gcp":
		return GardenerGcp, nil
	default:
		return UnknownGardener, nil
	}
}

func extractHyperScalerFromNode(node corev1.Node) Platform {
	if strings.HasPrefix(node.Name, "gke") {
		return Gke
	}

	if strings.HasPrefix(node.Name, "aks") {
		return Aks
	}

	return Unknown
}

func GetHyperScalerPlatform(clientset *kubernetes.Clientset) (Platform, error) {
	cm, err := clientset.CoreV1().ConfigMaps(shootCmNamespace).Get(shootCmName, metav1.GetOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return "", err
	}

	if err != nil && apierrors.IsNotFound(err) {
		nodes, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return "", err
		}

		return extractHyperScalerFromNode(nodes.Items[0]), nil
	}

	return extractHyperScalerFromCm(*cm)
}
