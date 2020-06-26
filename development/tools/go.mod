module github.com/kyma-project/test-infra/development/tools

go 1.13

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.2.0+incompatible
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
)

require (
	cloud.google.com/go/storage v1.6.0
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-querystring v1.0.0
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/jamiealquiza/envy v1.1.0
	github.com/kyma-project/test-infra/development/gcp v0.0.0-20200507124533-9b586ac404eb
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/githubv4 v0.0.0-20200414012201-bbc966b061dd
	github.com/sirupsen/logrus v1.5.0
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/tools v0.0.0-20200626032829-bcbc01e07a20 // indirect
	google.golang.org/api v0.22.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/test-infra v0.0.0-20200320172837-fbc86f22b087
)
