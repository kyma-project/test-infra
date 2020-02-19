package secrets

import (
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/config"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/storage"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func NewPopulator(kubeconfigPath string, kmsclient *Client, gcsclient *storage.Client) *Populator {
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	fatalOnError(err)
	k8sCli, err := kubernetes.NewForConfig(k8sConfig)
	fatalOnError(err)
	return &Populator{
		secretsClient: k8sCli.CoreV1().Secrets(metav1.NamespaceDefault),
		kmsClient: kmsclient,
		gcsClient: gcsclient,
	}
}

type Populator struct {
	secretsClient typedv1.SecretInterface
	kmsClient *Client
	gcsClient *storage.Client
}

// SecretModel defines secret to be stored in k8s
type SecretModel struct {
	Name string
	Key  string
}


//func (p *Populator) PopulateSaSecret(sakey *iam.ServiceAccountKey) error {
//	decodedkey, err := base64.StdEncoding.DecodeString(sakey)
//}
/*
func (p *Populator) PopulateAllFromConfig(config *config.Config) error {
	secrets := p.secretsFromConfig(config)

}
*/


func (p *Populator) saSecretsFromConfig(config *config.Config) []SecretModel {
	secretModel := make([]SecretModel, 0)
	for _, sa := range config.ServiceAccounts {
		secretModel = append(secretModel, SecretModel{
			Name: sa.Name,
			Key:  "service-account.json",
		})
	}
	return secretModel
}


func (p *Populator) genericSecretsFromConfig(config *config.Config) []SecretModel {
	for _, gen := range config.GenericSecrets {
		secretModel = append(secretModel, SecretModel{
			Name: gen.Name,
			Key:  gen.Key,
		})
	}
	return secretModel
}


func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}