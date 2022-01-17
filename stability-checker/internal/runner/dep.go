package runner

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type configMapClient interface {
	Create(context.Context, *v1.ConfigMap, metaV1.CreateOptions) (*v1.ConfigMap, error)
	Update(context.Context, *v1.ConfigMap, metaV1.UpdateOptions) (*v1.ConfigMap, error)
	Get(ctx context.Context, name string, options metaV1.GetOptions) (*v1.ConfigMap, error)
}
