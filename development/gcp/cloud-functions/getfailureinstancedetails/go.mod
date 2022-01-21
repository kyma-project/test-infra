module github.com/kyma-project/test-infra/development/gcp/cloud-functions/getfailureinstancedetails

go 1.14

replace (
	k8s.io/api => k8s.io/api v0.20.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.2
	k8s.io/client-go => k8s.io/client-go v0.20.2
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.2.0
)

require (
	cloud.google.com/go v0.88.0
	cloud.google.com/go/firestore v1.5.0
	cloud.google.com/go/pubsub v1.10.3
	github.com/google/go-github/v36 v36.0.0
	github.com/kyma-project/test-infra v0.0.0-20210826124132-fce6cc2cee02
)
