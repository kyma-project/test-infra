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
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/tidwall/gjson v1.9.3
)
