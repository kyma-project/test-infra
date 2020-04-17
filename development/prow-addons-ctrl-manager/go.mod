module github.com/kyma-project/test-infra/development/prow-addons-ctrl-manager

go 1.13

replace (
	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.5-beta.1
	k8s.io/client-go => k8s.io/client-go v0.16.4
)

require (
	github.com/go-logr/logr v0.1.0
	github.com/google/go-github v17.0.0+incompatible
	github.com/nlopes/slack v0.6.0
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/vektra/mockery v0.0.0-20181123154057-e78b021dcbb5
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/code-generator v0.17.3
	k8s.io/test-infra v0.0.0-20200417080107-13d4a722f14a
	sigs.k8s.io/controller-runtime v0.5.2
)
