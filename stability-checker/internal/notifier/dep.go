package notifier

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type configMapClient interface {
	Get(ctx context.Context, name string, options metaV1.GetOptions) (*v1.ConfigMap, error)
}
