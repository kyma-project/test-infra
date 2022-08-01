package secretversionsmanager

import (
	"github.com/kyma-project/test-infra/development/gcp/pkg/secretmanager"
	gcpsecretmanager "google.golang.org/api/secretmanager/v1"
)

func NewService(secretSvc *secretmanager.Service) *Service {
	secretVersionsService := gcpsecretmanager.NewProjectsSecretsVersionsService(secretSvc.Service)
	return &Service{ProjectsSecretsVersionsService: secretVersionsService}
}

func (svm *Service) DisableSecretVersion(version *gcpsecretmanager.SecretVersion) (*gcpsecretmanager.SecretVersion, error) {
	disableRequest := gcpsecretmanager.DisableSecretVersionRequest{}
	disableCall := svm.Disable(version.Name, &disableRequest)
	returnedSecretVersion, err := disableCall.Do()
	return returnedSecretVersion, err
}

func (svm *Service) GetSecretVersionData(secretPath string) (string, error) {
	secretVersionCall := svm.Access(secretPath)
	secretVersion, err := secretVersionCall.Do()
	if err != nil {
		return "", err
	}
	return secretVersion.Payload.Data, err
}
