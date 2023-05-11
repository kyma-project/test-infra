package extractimageurls

import (
	"reflect"
	"strings"
	"testing"
)

func TestFromKubernetesDeployments(t *testing.T) {
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
			ExpectedImages: nil,
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
		{
			Name:           "tekton deployment, pass",
			WantErr:        false,
			ExpectedImages: []string{"gcr.io/tekton-releases/github.com/tektoncd/dashboard/cmd/dashboard:v0.34.0@sha256:3b62e3d2423d28200f4125852c89b23a0725b89af355c4768d60d45d8c30fc47"},
			FileContent: `apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: tekton-dashboard
    app.kubernetes.io/component: dashboard
    app.kubernetes.io/instance: default
    app.kubernetes.io/name: dashboard
    app.kubernetes.io/part-of: tekton-dashboard
    app.kubernetes.io/version: v0.34
    dashboard.test.dev/release: v0.34
    version: v0.34.0
  name: tekton-dashboard
  namespace: tekton-pipelines
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/component: dashboard
      app.kubernetes.io/instance: default
      app.kubernetes.io/name: dashboard
      app.kubernetes.io/part-of: tekton-dashboard
  template:
    metadata:
      labels:
        app: tekton-dashboard
        app.kubernetes.io/component: dashboard
        app.kubernetes.io/instance: default
        app.kubernetes.io/name: dashboard
        app.kubernetes.io/part-of: tekton-dashboard
        app.kubernetes.io/version: v0.34.0
      name: tekton-dashboard
    spec:
      containers:
        - args:
            - --port=1000
          image: gcr.io/tekton-releases/github.com/tektoncd/dashboard/cmd/dashboard:v0.34.0@sha256:3b62e3d2423d28200f4125852c89b23a0725b89af355c4768d60d45d8c30fc47
`,
		},
		{
			Name:           "complex file with multiple resources, pass",
			WantErr:        false,
			ExpectedImages: []string{"test-image:test"},
			FileContent: `apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: default
  name: "test"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  labels:
    app: test
spec:
  replicas: 3
  selector:
    matchLabels:
       app: test
  template:
    metadata:
      labels:
          app: test
    spec:
      containers:
        - name: test
          image: test-image:test`,
		},
		{
			Name:           "complex file with multiple deployments, pass",
			WantErr:        false,
			ExpectedImages: []string{"test-image:test", "test-image:test2"},
			FileContent: `apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: default
  name: "test"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  labels:
    app: test
spec:
  replicas: 3
  selector:
    matchLabels:
       app: test
  template:
    metadata:
      labels:
          app: test
    spec:
      containers:
        - name: test
          image: test-image:test
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test2-deployment
  labels:
    app: test2
spec:
  replicas: 3
  selector:
    matchLabels:
       app: test2
  template:
    metadata:
      labels:
          app: test2
    spec:
      containers:
        - name: test2
          image: test-image:test2`,
		},
		{
			Name:           "deployment with resource quantity, pass",
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
              protocol: TCP
          resources:
            requests:
              cpu: 100m
              memory: 100Mi`,
		},
	}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			images, err := FromKubernetesDeployments(strings.NewReader(c.FileContent))
			if err != nil && !c.WantErr {
				t.Errorf("Got unexpected error: %v", err)
			}

			if !reflect.DeepEqual(images, c.ExpectedImages) {
				t.Errorf("FromKubernetesDeployments(): Got %v, but expected %v", images, c.ExpectedImages)
			}
		})
	}
}
