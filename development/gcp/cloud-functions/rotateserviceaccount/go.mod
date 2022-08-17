module github.com/kyma-project/test-infra/development/gcp/cloud-functions/rotateserviceaccount

go 1.16

replace (
	k8s.io/api => k8s.io/api v0.22.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
	k8s.io/client-go => k8s.io/client-go v0.22.2
)

require (
	cloud.google.com/go/compute v1.7.0
	github.com/kyma-project/test-infra v0.0.0-20220817091107-5daa3eca1059
	google.golang.org/api v0.92.0
)
