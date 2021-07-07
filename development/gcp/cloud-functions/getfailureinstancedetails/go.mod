module github.com/kyma-project/test-infra/development/gcp/cloud-functions/getfailureinstancedetails

go 1.14

replace github.com/kyma-project/test-infra v0.0.0 => /Users/i319037/go/src/github.com/kyma-project/test-infra

require (
	cloud.google.com/go v0.81.0
	cloud.google.com/go/firestore v1.1.0
	cloud.google.com/go/pubsub v1.4.0
	github.com/google/go-cmp v0.5.6 // indirect
	google.golang.org/api v0.46.0 // indirect
	k8s.io/test-infra v0.0.0-20210707063243-a1086c936604
)
