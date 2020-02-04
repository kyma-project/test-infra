package roles

import (
	"context"
	"fmt"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

type crmService struct {
	credentials string
	ctx         context.Context
	service     *cloudresourcemanager.Service
}

func NewService(credentials string) (*crmService, error) {
	ctx := context.Background()
	service, err := cloudresourcemanager.NewService(ctx, option.WithCredentialsFile(credentials))
	if err != nil {
		return nil, fmt.Errorf("When creating new GCP cloudresourcemanager service got error: [%v].", err)
	} else {
		return &crmService{
			credentials: credentials,
			ctx:         ctx,
			service:     service,
		}, err
	}
}

func (crms *crmService) GetPolicy(projectname string, getiampolicyrequest *cloudresourcemanager.GetIamPolicyRequest) (*cloudresourcemanager.Policy, error) {
	return crms.service.Projects.GetIamPolicy(projectname, getiampolicyrequest).Context(crms.ctx).Do()
}

func (crms *crmService) SetPolicy(projectname string, setiampolicyrequest *cloudresourcemanager.SetIamPolicyRequest) (*cloudresourcemanager.Policy, error) {
	return crms.service.Projects.SetIamPolicy(projectname, setiampolicyrequest).Context(crms.ctx).Do()
}
