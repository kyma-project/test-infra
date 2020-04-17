IMG_NAME = kyma-integration
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

ci-pr: build-image push-image
ci-release: build-image push-image

build-image:
	docker build -t $(IMG):$(TAG)-k8s1.16 --build-arg K8S_VERSION=1.16 .
	docker build -t $(IMG):$(TAG)-k8s1.15 --build-arg K8S_VERSION=1.15 .
	docker build -t $(IMG):$(TAG)-k8s1.14 --build-arg K8S_VERSION=1.14 .
push-image:

	docker push $(IMG) # it pushes all tags for defined image
