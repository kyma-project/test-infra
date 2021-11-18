module github.com/kyma-project/test-infra

go 1.16

replace (
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
)

require (
	cloud.google.com/go v0.88.0
	cloud.google.com/go/bigquery v1.8.0
	cloud.google.com/go/firestore v1.5.0
	cloud.google.com/go/logging v1.0.0
	cloud.google.com/go/pubsub v1.10.3
	cloud.google.com/go/storage v1.16.0
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/forestgiant/sliceutil v0.0.0-20160425183142-94783f95db6c
	github.com/ghodss/yaml v1.0.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/google/go-containerregistry v0.5.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v31 v31.0.0
	github.com/google/go-github/v36 v36.0.0
	github.com/google/go-github/v40 v40.0.0
	github.com/google/go-querystring v1.1.0
	github.com/imdario/mergo v0.3.11
	github.com/jamiealquiza/envy v1.1.0
	github.com/jinzhu/copier v0.3.2
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0
	github.com/shurcooL/githubv4 v0.0.0-20191102174205-af46314aec7b
	github.com/sigstore/cosign v1.0.1
	github.com/sigstore/sigstore v0.0.0-20210729211320-56a91f560f44
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.7.5
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	google.golang.org/api v0.50.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20210407040951-51f95c2d525e
)
