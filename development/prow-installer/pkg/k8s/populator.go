package k8s

import (
	"bytes"
	"encoding/base64"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/config"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func NewPopulator(k8sclient *kubernetes.Clientset) *Populator {
	return &Populator{
		//ctx: ctx,
		k8sClient: k8sclient,
		//secretsClient: k8sclient.CoreV1().Secrets(metav1.NamespaceDefault),
	}
}

type Populator struct {
	//ctx context.Context
	secrets []SecretModel
	k8sClient *kubernetes.Clientset
	secretsClient typedv1.SecretInterface
}

// SecretModel defines secret to be stored in k8s
type SecretModel struct {
	Name string
	Key  string
	KeyData string
}


//func (p *Populator) PopulateSaSecret(sakey *iam.ServiceAccountKey) error {
//	decodedkey, err := base64.StdEncoding.DecodeString(sakey)
//}


func (p *Populator) newSecretsClient(namespace string) {
	p.secretsClient = p.k8sClient.CoreV1().Secrets(namespace)
}
func (p *Populator) PopulateSecrets(secrets []SecretModel, namespace string) error {
	p.newSecretsClient(namespace)
	for _, secret := range secrets {
		curr, err := p.secretsClient.Get(secret.Name, metav1.GetOptions{})
		decoded, err := base64.StdEncoding.DecodeString(secret.KeyData)
		switch {
		case err == nil:
			if bytes.Equal(curr.Data[secret.Key], decoded) {
				s.logSecretAction(curr, "Unchanged")
				continue
			}

			curr.Data[sec.Key] = decoded
			if _, err = s.secretsClient.Update(curr); err != nil {
				return errors.Wrap(err, "while updating secret")
			}
			s.logSecretAction(curr, "Updated")

		case k8serrors.IsNotFound(err):
			curr, err := s.secretsClient.Create(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: sec.Name,
				},
				Data: map[string][]byte{
					sec.Key: decoded,
				},
			})
			if err != nil {
				return errors.Wrapf(err, "while creating secret [%s]", sec.Name)
			}
			s.logSecretAction(curr, "Created")
		default:
			return errors.Wrapf(err, "while getting secret [%s]", sec.Name)
		}
	}
}



func (p *Populator) saSecretsFromConfig(config *config.Config) []SecretModel {
	secretModel := make([]SecretModel, 0)
	for _, sa := range config.ServiceAccounts {
		secretModel = append(secretModel, SecretModel{
			Name: sa.Name,
			Key:  "service-account.json",
			KeyData: sa.Key,
		})
	}
	return secretModel
}


func (p *Populator) genericSecretsFromConfig(config *config.Config) []SecretModel {
	var secretModel []SecretModel
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