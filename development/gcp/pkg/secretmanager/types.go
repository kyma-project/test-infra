package secretmanager

import (
	secretmanager "google.golang.org/api/secretmanager/v1"
)

type Service struct {
	Service        *secretmanager.Service
	VersionService *secretmanager.ProjectsSecretsVersionsService
}
