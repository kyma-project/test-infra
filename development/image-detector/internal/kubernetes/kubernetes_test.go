package kubernetes

import (
	"reflect"
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {
	tc := []struct {
		Name           string
		FileContent    string
		WantErr        bool
		ExpectedImages []string
	}{
		{
			Name:           "simple prow component deployment, pass",
			WantErr:        false,
			ExpectedImages: []string{"test.gcr.io/test:test"},
			FileContent: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test
  labels:
    app: test
spec:
  replica: 2
  selector:
    matchLabels:
      app: test
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: test
    spec:
      terminationGracePeriodSeconds: 30
      containers:
        - name: test
          image: test.gcr.io/test:test
          args:
            - -tets-arg
          ports:
            - containerPort: 8080
              protocol: TCP`,
		},
		{
			Name:           "valid service file, fail",
			WantErr:        true,
			ExpectedImages: []string{},
			FileContent: `apiVersion: v1
kind: Service
metadata:
  metadata:
    name: test
    labels:
      apps: test
spec:
  selector:
    app: test
    type: NodePort
    ports:
      - name: http
        port: 80
        targetPort: 8080`,
		},
	}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			images, err := extract(strings.NewReader(c.FileContent))
			if err != nil && !c.WantErr {
				t.Errorf("Got unexpected error: %v", err)
			}

			if !reflect.DeepEqual(images, c.ExpectedImages) {
				t.Errorf("%v != %v", images, c.ExpectedImages)
			}
		})
	}
}
