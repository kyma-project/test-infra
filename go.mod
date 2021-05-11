module github.com/kyma-project/test-infra

go 1.16

replace (
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
)

require (
	github.com/containerd/containerd v1.5.0 // indirect
	github.com/docker/docker v20.10.6+incompatible
	github.com/elithrar/admission-control v0.6.7
	github.com/go-kit/kit v0.10.0
	github.com/google/gofuzz v1.2.1-0.20210504230335-f78f29fc09ea // indirect
	github.com/gorilla/mux v1.8.0
	github.com/jamiealquiza/envy v1.1.0
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/novln/docker-parser v1.0.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	google.golang.org/api v0.46.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20210407040951-51f95c2d525e
)
