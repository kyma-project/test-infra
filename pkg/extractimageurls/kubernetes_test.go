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
			Name:           "valid service file, pass",
			WantErr:        false,
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
		{
			Name:           "config map with yaml file, pass",
			WantErr:        false,
			ExpectedImages: nil,
			FileContent: `apiVersion: v1
kind: ConfigMap
metadata:
  namespace: prow-monitoring
  name: grafana-datasources
data:
  datasources.yaml: |
    ---
    apiVersion: 1
    datasources:
      - name: prometheus
        type: prometheus
        access: proxy
        url: http://prometheus.prow-monitoring.svc:9090
        version: 1
        orgId: 1
        editable: false
      - name: bigquery-prow
        type: doitintl-bigquery-datasource
        access: proxy
        jsonData:
            authenticationType: jwt
            clientEmail: ${GF_PLUGINS_BIGQUERY_DATASOURCE_EMAIL}
            defaultProject: sap-kyma-prow
            tokenUri: https://oauth2.googleapis.com/token
        secureJsonData:
            privateKey: "${GF_PLUGINS_BIGQUERY_DATASOURCE_PRIVATE_KEY}"
        version: 2
        readOnly: false`,
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
# (2025-03-04)