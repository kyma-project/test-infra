package resource

import (
	"k8s.io/apimachinery/pkg/labels"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
)

type Resource struct {
	ResCli    dynamic.ResourceInterface
	namespace string
	kind      string
}

func New(dynamicCli dynamic.Interface, s schema.GroupVersionResource, namespace string) *Resource {
	resCli := dynamicCli.Resource(s).Namespace(namespace)
	return &Resource{ResCli: resCli, namespace: namespace, kind: s.Resource}
}

func (r *Resource) List(set map[string]string) (*unstructured.UnstructuredList, error) {
	var result *unstructured.UnstructuredList

	selector := labels.SelectorFromSet(set).String()
	err := retry.OnError(retry.DefaultBackoff, func(err error) bool {
		if apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) || apierrors.IsTooManyRequests(err) {
			return true
		}
		return false
	}, func() error {
		var listErr error
		result, listErr = r.ResCli.List(metav1.ListOptions{
			LabelSelector: selector,
		})
		return listErr
	})
	if err != nil {
		return nil, errors.Wrapf(err, "while listing resource %s in namespace %s", r.kind, r.namespace)
	}

	return result, nil
}
