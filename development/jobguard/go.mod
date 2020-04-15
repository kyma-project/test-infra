module github.com/kyma-project/test-infra/development/jobguard

go 1.13

replace (
	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.5-beta.1
	k8s.io/client-go => k8s.io/client-go v0.16.4
)

require (
	github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/kyma-project/test-infra v0.0.0-20200331110003-6fa7a9f1e555
	github.com/pkg/errors v0.9.1
	github.com/vrischmann/envconfig v1.2.0
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/test-infra v0.0.0-20200331085241-bf7dc9346358
)
