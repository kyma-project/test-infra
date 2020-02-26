package k8s

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/kyma-project/test-infra/development/prow-installer/pkg/serviceaccount"
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
	secrets       []SecretModel
	k8sClient     *kubernetes.Clientset
	secretsClient typedv1.SecretInterface
}

// SecretModel defines secret to be stored in k8s
type SecretModel struct {
	Name    string
	Key     string
	KeyData string
}

type GenericSecret struct {
	Name string `yaml:"prefix"`
	Key  string `yaml:"key"`
}

func (p *Populator) newSecretsClient(namespace string) {
	p.secretsClient = p.k8sClient.CoreV1().Secrets(namespace)
}
func (p *Populator) PopulateSecrets(namespace string, generics []GenericSecret, sasecrets []serviceaccount.ServiceAccount) error {
	var secrets []SecretModel
	p.newSecretsClient(namespace)
	secrets = p.saSecretsFromConfig(secrets, sasecrets)
	secrets = p.genericSecretsFromConfig(secrets, generics)
	for _, secret := range secrets {
		decoded, err := base64.StdEncoding.DecodeString(secret.KeyData)
		if err != nil {
			return fmt.Errorf("failed get secret %s from config, got: %v", secret.Name, err)
		}
		curr, err := p.secretsClient.Get(secret.Name, metav1.GetOptions{})
		switch {
		case err == nil:
			if bytes.Equal(curr.Data[secret.Key], decoded) {
				continue
			}
			curr.Data[secret.Key] = decoded
			if _, err = p.secretsClient.Update(curr); err != nil {
				return fmt.Errorf("while updating secret, got error: %w", err)
			}

		case k8serrors.IsNotFound(err):
			curr, err := p.secretsClient.Create(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: secret.Name,
				},
				Data: map[string][]byte{
					secret.Key: decoded,
				},
			})
			if err != nil {
				return fmt.Errorf("while creating secret %s, got error: %w", secret.Name, err)
			}
			log.Printf("loaded secret %s", curr.Name)
		default:
			return fmt.Errorf("while getting secret %s, got error: %w", secret.Name, err)
		}
	}
	return nil
}

func (p *Populator) saSecretsFromConfig(secrets []SecretModel, saSecrets []serviceaccount.ServiceAccount) []SecretModel {
	for _, sa := range saSecrets {
		secrets = append(secrets, SecretModel{
			Name:    sa.Name,
			Key:     "service-account.json",
			KeyData: sa.Key.PrivateKeyData,
		})
	}
	return secrets
}

func (p *Populator) genericSecretsFromConfig(secrets []SecretModel, generics []GenericSecret) []SecretModel {
	for _, gen := range generics {
		secrets = append(secrets, SecretModel{
			Name: gen.Name,
			Key:  gen.Key,
		})
	}
	return secrets
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
