module github.com/kyma-project/test-infra

go 1.16

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.9
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/opencontainers/distribution-spec => github.com/opencontainers/distribution-spec v1.0.1
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	// these two sigstore/* replaces are required for the image-syncer tool
	github.com/sigstore/cosign => github.com/sigstore/cosign v1.2.1
	github.com/sigstore/sigstore => github.com/sigstore/sigstore v1.0.1
	k8s.io/api => k8s.io/api v0.22.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
	k8s.io/client-go => k8s.io/client-go v0.22.2
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.9.0
)

require (
	cloud.google.com/go/bigquery v1.30.0
	cloud.google.com/go/firestore v1.6.1
	cloud.google.com/go/functions v1.3.0
	cloud.google.com/go/logging v1.4.2
	cloud.google.com/go/pubsub v1.19.0
	cloud.google.com/go/storage v1.22.0
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/containerd/containerd v1.6.1
	github.com/containerd/typeurl v1.0.2
	github.com/docker/docker v20.10.12+incompatible // indirect
	github.com/forestgiant/sliceutil v0.0.0-20160425183142-94783f95db6c
	github.com/fsouza/go-dockerclient v1.7.10
	github.com/gardener/component-cli v0.38.0
	github.com/gardener/component-spec/bindings-go v0.0.59
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.4.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/google/go-containerregistry v0.7.1-0.20211203164431-c75901cce627
	github.com/google/go-github/v40 v40.0.0
	github.com/google/go-github/v42 v42.0.0
	github.com/google/go-querystring v1.1.0
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12
	github.com/jamiealquiza/envy v1.1.0
	github.com/jinzhu/copier v0.3.5
	github.com/mandelsoft/vfs v0.0.0-20210530103237-5249dc39ce91
	github.com/onsi/gomega v1.19.0
	github.com/opencontainers/runc v1.0.3 // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/shurcooL/githubv4 v0.0.0-20211117020012-5800b9de5b8b
	github.com/sigstore/cosign v1.2.1
	github.com/sigstore/sigstore v1.0.1
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/goconvey v1.7.2
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.0 // indirect
	github.com/stretchr/testify v1.7.1
	github.com/tidwall/gjson v1.14.0
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/zap v1.21.0
	golang.org/x/oauth2 v0.0.0-20220309155454-6242fa91716a
	google.golang.org/api v0.74.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20211214153147-7b2c2d007b33
	sigs.k8s.io/yaml v1.3.0
)
