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
	cloud.google.com/go/pubsub v1.10.3
	cloud.google.com/go/storage v1.16.0
	contrib.go.opencensus.io/exporter/stackdriver v0.13.8 // indirect
	github.com/Azure/go-autorest/autorest v0.11.19 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.14 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/aws/aws-sdk-go v1.40.7 // indirect
	github.com/fatih/color v1.12.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/forestgiant/sliceutil v0.0.0-20160425183142-94783f95db6c
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/google/btree v1.0.1 // indirect
	github.com/google/go-containerregistry v0.5.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v31 v31.0.0
	github.com/google/go-github/v36 v36.0.0
	github.com/google/go-querystring v1.0.0
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v0.16.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.0 // indirect
	github.com/imdario/mergo v0.3.11
	github.com/jamiealquiza/envy v1.1.0
	github.com/jinzhu/copier v0.3.2
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.29.0 // indirect
	github.com/prometheus/procfs v0.7.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/shurcooL/githubv4 v0.0.0-20191102174205-af46314aec7b
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.7.5
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/vrischmann/envconfig v1.3.0
	go.uber.org/atomic v1.8.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.18.1 // indirect
	gocloud.dev v0.23.0 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210726213435-c6fcb2dbf985 // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
	google.golang.org/api v0.50.0
	google.golang.org/genproto v0.0.0-20210722135532-667f2b7c528f // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/test-infra v0.0.0-20210407040951-51f95c2d525e
)
