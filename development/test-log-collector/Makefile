APP_NAME = test-log-collector
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: test
test:
	go test ./... -count=1

.PHONY: build-image
build-image:
	docker build -t $(APP_NAME):latest .

.PHONY: push-image
push-image:
	docker tag $(APP_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)

.PHONY: ci-pr
ci-pr: test build-image push-image

.PHONY: ci-master
ci-master: test build-image push-image

.PHONY: ci-release
ci-release: test build-image push-image

