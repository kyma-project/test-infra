module github.com/kyma-project/test-infra/development/prow-installer

go 1.13

require (
	cloud.google.com/go v0.49.0
	cloud.google.com/go/storage v1.4.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/google/go-cmp v0.3.0
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	google.golang.org/api v0.14.0
	google.golang.org/genproto v0.0.0-20191115221424-83cc0476cb11
	gopkg.in/yaml.v2 v2.2.8
	//k8s.io/api v0.17.3 // indirect
	k8s.io/api v0.17.2
	//k8s.io/apimachinery v0.17.3
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/utils v0.0.0-20200124190032-861946025e34 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
