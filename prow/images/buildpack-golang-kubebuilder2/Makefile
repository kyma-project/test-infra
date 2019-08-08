IMG_NAME ?= buildpack-golang-kubebuilder2
DOCKER_PUSH_REPOSITORY ?= eu.gcr.io/kyma-project
DOCKER_TAG ?= latest

IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)


ci-pr: build-image push-image
ci-release: build-image push-image

build-image:
	docker build --build-arg BUILDPACK_VERSION=$(DOCKER_TAG) -t $(IMG):$(DOCKER_TAG) .
push-image:
	docker push $(IMG):$(DOCKER_TAG)
