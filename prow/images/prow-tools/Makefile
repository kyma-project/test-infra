IMG_NAME = prow-tools
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

.PHONY: build-image
build-image:
	docker build -t $(IMG_NAME) --build-arg commit=$(TAG) -f Dockerfile ../../../.

.PHONY: push-image
push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

.PHONY: ci-pr
ci-pr: build-image push-image

.PHONY: ci-release
ci-release: build-image push-image
	docker tag $(IMG_NAME) $(IMG):current
	docker push $(IMG):current

.PHONY: clean
clean:
	rm -f $(APP_NAME)
