module github.com/kyma-project/test-infra/development/prow-installer

go 1.13

require (
	cloud.google.com/go v0.54.0
	cloud.google.com/go/storage v1.6.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.2
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.6.1
	google.golang.org/api v0.20.0
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.20.0
)
