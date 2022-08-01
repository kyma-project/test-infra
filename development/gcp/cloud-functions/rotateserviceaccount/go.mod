module github.com/kyma-project/test-infra/development/gcp/cloud-functions/rotateserviceaccount

go 1.16

require github.com/kyma-project/test-infra v0.0.0-20220715122928-d02a288f4078

require (
	cloud.google.com/go v0.102.1
	cloud.google.com/go/iam v0.3.0
	cloud.google.com/go/pubsub v1.23.1
	github.com/beorn7/perks v1.0.1
	github.com/cespare/xxhash/v2 v2.1.2
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/evanphx/json-patch/v5 v5.5.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v1.2.2
	github.com/gogo/protobuf v1.3.2
	github.com/gomodule/redigo v1.8.5
	github.com/google/btree v1.0.1
	github.com/google/go-cmp v0.5.8
	github.com/google/go-github/v42 v42.0.0
	github.com/google/go-querystring v1.1.0
	github.com/google/gofuzz v1.2.1-0.20210504230335-f78f29fc09ea
	github.com/googleapis/gnostic v0.5.5
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc
	github.com/json-iterator/go v1.1.12
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.2
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.30.0
	github.com/prometheus/procfs v0.7.3
	github.com/shurcooL/githubv4 v0.0.0-20211117020012-5800b9de5b8b
	github.com/shurcooL/graphql v0.0.0-20181231061246-d48a9a75455f
	github.com/sirupsen/logrus v1.9.0
	github.com/tektoncd/pipeline v0.14.1-0.20200710073957-5eeb17f81999
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	golang.org/x/term v0.0.0-20220526004731-065cf7ba2467
	golang.org/x/time v0.0.0-20220609170525-579cf78fd858
	gomodules.xyz/jsonpatch/v2 v2.2.0
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/klog/v2 v2.30.0
	k8s.io/kube-openapi v0.0.0-20211109043538-20434351676c
	k8s.io/test-infra v0.0.0-20211214153147-7b2c2d007b33
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
	knative.dev/pkg v0.0.0-20210908025933-71508fc69a57
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go/compute v1.7.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/googleapis/enterprise-certificate-proxy v0.1.0
	github.com/googleapis/gax-go/v2 v2.4.0
	go.opencensus.io v0.23.0
	go.uber.org/atomic v1.9.0
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.21.0
	golang.org/x/net v0.0.0-20220624214902-1bab6f366d9e
	golang.org/x/oauth2 v0.0.0-20220622183110-fd043fe589d2
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8
	golang.org/x/text v0.3.7
	google.golang.org/api v0.88.0
	google.golang.org/appengine v1.6.7
	google.golang.org/genproto v0.0.0-20220714211235-042d03aeabc9
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.28.0
	k8s.io/api v0.22.5
	k8s.io/apimachinery v0.22.5
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
)

replace (
	github.com/kyma-project/test-infra v0.0.0-20220715122928-d02a288f4078 => /Users/i542853/go/src/github.com/kyma-project/test-infra
	k8s.io/api => k8s.io/api v0.22.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
	k8s.io/client-go => k8s.io/client-go v0.22.2
)
