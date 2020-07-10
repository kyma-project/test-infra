IMG_NAME = probot-stale
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

ci-pr: build-image push-image
ci-release: build-image push-image

build-image:
	docker build -t $(IMG):$(TAG) .
push-image:
	docker push $(IMG):$(TAG)
