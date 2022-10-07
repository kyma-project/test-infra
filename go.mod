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
	cloud.google.com/go/bigquery v1.42.0
	cloud.google.com/go/compute v1.10.0
	cloud.google.com/go/firestore v1.7.0
	cloud.google.com/go/functions v1.7.0
	cloud.google.com/go/logging v1.5.0
	cloud.google.com/go/pubsub v1.25.1
	cloud.google.com/go/storage v1.27.0
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/cloudevents/sdk-go/v2 v2.12.0
	github.com/containerd/containerd v1.6.6
	github.com/containerd/typeurl v1.0.2
	github.com/forestgiant/sliceutil v0.0.0-20160425183142-94783f95db6c
	github.com/fsouza/go-dockerclient v1.8.3
	github.com/ghodss/yaml v1.0.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/google/go-containerregistry v0.7.1-0.20211203164431-c75901cce627
	github.com/google/go-github/v40 v40.0.0
	github.com/google/go-github/v42 v42.0.0
	github.com/google/go-querystring v1.1.0
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.13
	github.com/jamiealquiza/envy v1.1.0
	github.com/jinzhu/copier v0.3.5
	github.com/onsi/gomega v1.20.2
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/githubv4 v0.0.0-20211117020012-5800b9de5b8b
	github.com/sigstore/cosign v1.2.1
	github.com/sigstore/sigstore v1.0.1
	github.com/sirupsen/logrus v1.9.0
	github.com/smartystreets/goconvey v1.7.2
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.0 // indirect
	github.com/stretchr/testify v1.8.0
	github.com/tidwall/gjson v1.14.3
	go.uber.org/zap v1.23.0
	golang.org/x/net v0.0.0-20220909164309-bea034e7d591
	golang.org/x/oauth2 v0.0.0-20220909003341-f21342109be1
	google.golang.org/api v0.98.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20211214153147-7b2c2d007b33
	sigs.k8s.io/yaml v1.3.0
)
