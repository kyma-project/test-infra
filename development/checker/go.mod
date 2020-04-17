module github.com/kyma-project/test-infra/development/checker

go 1.13

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
	k8s.io/code-generator => k8s.io/code-generator v0.17.3
)

require (
	github.com/aws/aws-sdk-go v1.29.32 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/sirupsen/logrus v1.5.0
	go.uber.org/zap v1.14.1 // indirect
	google.golang.org/grpc v1.27.0 // indirect
	k8s.io/apimachinery v0.17.3
	k8s.io/test-infra v0.0.0-20200320172837-fbc86f22b087
	sigs.k8s.io/yaml v1.1.0
)
