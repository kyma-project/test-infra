module github.com/kyma-project/test-infra/development/test-log-collector

go 1.14

require (
	cloud.google.com/go v0.57.0 // indirect
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/slack-go/slack v0.6.5
	github.com/vrischmann/envconfig v1.2.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.17.13
	k8s.io/apimachinery v0.17.13
	k8s.io/client-go v0.17.13
	sigs.k8s.io/controller-runtime v0.5.11
)
