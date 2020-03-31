IMG_NAME = kyma-integration
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

ci-pr: build-image push-image
ci-release: build-image push-image

build-image:
	docker build -t $(IMG_NAME) .
	docker tag $(IMG_NAME) $(IMG):$(TAG)
push-image:
	docker tag $(IMG_NAME) $(IMG):k8s-1.16
	docker push $(IMG):$(TAG)
