module github.com/kyma-project/test-infra/development/gcp/cloud-functions/getgithubissue

go 1.14

replace (
	k8s.io/api => k8s.io/api v0.21.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.1
	k8s.io/client-go => k8s.io/client-go v0.21.1
)

require (
	cloud.google.com/go v0.89.0
	cloud.google.com/go/firestore v1.5.0
	cloud.google.com/go/pubsub v1.13.0
	github.com/google/go-github/v36 v36.0.0
	github.com/kyma-project/test-infra v0.0.0-20210826124132-fce6cc2cee02
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
)
