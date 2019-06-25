APP_NAME ?= prow-addons-ctrl-manager
IMG ?= $(APP_NAME):latest
IMG-CI = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME):$(DOCKER_TAG)

MOCKERY= go run ./vendor/github.com/vektra/mockery/cmd/mockery/mockery.go

all: test manager

# Run tests
test: generate manifests validate
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate validate
	go build -o bin/manager github.com/kyma-project/test-infra/development/prow-addons-ctrl-manager/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kubectl apply -f config/deploy

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all
	kustomize build config > config/deploy/deploy.yaml

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...
	${MOCKERY} -dir=./pkg/controller/notifier/ -name=ReportClient -output=./pkg/controller/notifier/automock -outpkg=automock -case=underscore

# Build the docker image
docker-build: resolve test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker tag $(IMG) $(IMG-CI)
	docker push $(IMG-CI)

### Custom targets
# Resolve dependencies
resolve:
	dep ensure -v -vendor-only

validate: fmt vet
	dep status

# CI specified targets
ci-pr: docker-build docker-push
ci-master: docker-build docker-push
ci-release: docker-build docker-push
