module github.com/kyma-project/test-infra

go 1.16

replace (
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2

)

require (
	cloud.google.com/go v0.81.0
	cloud.google.com/go/bigquery v1.8.0
	cloud.google.com/go/pubsub v1.4.0
	cloud.google.com/go/storage v1.12.0
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/containerd/containerd v1.5.0 // indirect
	github.com/docker/docker v20.10.6+incompatible
	github.com/forestgiant/sliceutil v0.0.0-20160425183142-94783f95db6c
	github.com/ghodss/yaml v1.0.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v31 v31.0.0
	github.com/google/go-github/v36 v36.0.0
	github.com/google/go-querystring v1.0.0
	github.com/google/gofuzz v1.2.1-0.20210504230335-f78f29fc09ea // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/imdario/mergo v0.3.11
	github.com/jamiealquiza/envy v1.1.0
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/novln/docker-parser v1.0.0
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/shurcooL/githubv4 v0.0.0-20191102174205-af46314aec7b
	github.com/sirupsen/logrus v1.7.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/tidwall/gjson v1.7.5
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/oauth2 v0.0.0-20210427180440-81ed05c6b58c
	google.golang.org/api v0.46.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20210407040951-51f95c2d525e
	sigs.k8s.io/structured-merge-diff/v4 v4.1.1 // indirect
)
