package secretmanager

import (
	secretmanager "google.golang.org/api/secretmanager/v1"
)

// Service contains services for secret manipulation
type Service struct {
	Service        *secretmanager.Service
	VersionService *secretmanager.ProjectsSecretsVersionsService
}
