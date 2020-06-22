module github.com/dekiel/test-infra/development/prowjob_parser

go 1.14

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
)

require (
	github.com/jamiealquiza/envy v1.1.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	golang.org/x/tools v0.0.0-20200619210111-0f592d2728bb // indirect
	k8s.io/test-infra v0.0.0-20200320172837-fbc86f22b087
)
