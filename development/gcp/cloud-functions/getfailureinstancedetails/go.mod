module github.com/kyma-project/test-infra/development/gcp/cloud-functions/getfailureinstancedetails

go 1.16

replace (
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.2.0
	k8s.io/api => k8s.io/api v0.23.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.3
	k8s.io/client-go => k8s.io/client-go v0.23.3
)

require (
	cloud.google.com/go/firestore v1.6.1
	cloud.google.com/go/functions v1.1.0
	cloud.google.com/go/pubsub v1.17.1
	github.com/google/go-github/v42 v42.0.0
	github.com/kyma-project/test-infra v0.0.0-20220207154300-3bedf56a4d7f
)
