APP_NAME = perf-test
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: build-image
build-image:
    $(info ************  BUILD IMAGE ************)
    docker build -f Dockerfile -t $(APP_NAME):latest ../../

.PHONY: push-image
push-image:
	docker tag $(APP_NAME):latest $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

.PHONY: ci-pr
ci-pr: build-image push-image

.PHONY: ci-master
ci-master: build-image push-image

.PHONY: ci-release
ci-release: build-image push-image