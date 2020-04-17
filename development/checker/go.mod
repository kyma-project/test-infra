module github.com/kyma-project/test-infra/development/checker

go 1.13

//cloud.google.com/go => cloud.google.com/go v0.44.3
//github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
//golang.org/x/lint => golang.org/x/lint v0.0.0-20190409202823-959b441ac422
replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
	k8s.io/code-generator => k8s.io/code-generator v0.17.3
)

require (
	github.com/sirupsen/logrus v1.5.0
	k8s.io/test-infra v0.0.0-20200331085241-bf7dc9346358
)
