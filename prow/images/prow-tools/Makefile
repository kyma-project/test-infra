IMG_NAME = prow-tools
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

resolve:
	go mod verify

validate: resolve
	./before-commit.sh

jobs-tests: resolve
	go test ./jobs/...

ci-pr: build-image push-image
ci-release: build-image push-image

build-image:
	docker build -t $(IMG_NAME) -f Dockerfile ../../../.
push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)
