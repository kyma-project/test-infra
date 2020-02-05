package serviceaccount

import (
	"context"
	"fmt"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

type IAMService struct {
	credentials string
	ctx         context.Context
	service     *iam.Service
}

func NewService(credentials string) (*IAMService, error) {
	var iamservice IAMService
	ctx := context.Background()
	service, err := iam.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, fmt.Errorf("Error %w when creating new IAMService.", err)
	} else {
		iamservice = IAMService{
			credentials: credentials,
			ctx:         ctx,
			service:     service,
		}
		return &iamservice, nil
	}
}

func (iams *IAMService) CreateSA(request *iam.CreateServiceAccountRequest, projectname string) (*iam.ServiceAccount, error) {
	return iams.service.Projects.ServiceAccounts.Create(projectname, request).Context(iams.ctx).Do()
}
