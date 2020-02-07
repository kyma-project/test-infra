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

func NewService(credentials string) (IAMService, error) {
	var iamservice IAMService
	ctx := context.Background()
	service, err := iam.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return iamservice, fmt.Errorf("Error %v when creating new IAMService.", err)
	} else {
		iamservice = IAMService{
			credentials: credentials,
			ctx:         ctx,
			service:     service,
		}
		return iamservice, nil
	}
}

func (iams *IAMService) CreateSA(saname string, projectname string) (*iam.ServiceAccount, error) {
	createsaaccountrequest := iam.CreateServiceAccountRequest{
		AccountId: saname,
	}
	return iams.service.Projects.ServiceAccounts.Create(projectname, &createsaaccountrequest).Context(iams.ctx).Do()
}
