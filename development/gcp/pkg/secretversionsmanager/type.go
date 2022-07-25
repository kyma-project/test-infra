package secretversionsmanager

import (
	secretmanager "google.golang.org/api/secretmanager/v1"
)

type Service struct {
	*secretmanager.ProjectsSecretsVersionsService
}
