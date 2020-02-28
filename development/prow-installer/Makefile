APP_NAME ?= prow-installer
IMG ?= $(APP_NAME):latest
IMG-CI = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME):$(DOCKER_TAG)

all: test installer

# Run tests
test:
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build config binary
installer: test
	go build -o bin/installer github.com/kyma-project/test-infra/development/prow-installer/cmd/installer

# Run against the configured Kubernetes cluster in ~/.kube/config
run:
	go run ./cmd/installer/main.go

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker tag $(IMG) $(IMG-CI)
	docker push $(IMG-CI)

# CI specified targets
ci-pr: docker-build docker-push
ci-master: docker-build docker-push
ci-release: docker-build docker-push
