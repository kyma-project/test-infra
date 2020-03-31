IMG_NAME = clonerefs
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

resolve:
	dep ensure -v -vendor-only

validate: resolve
	./before-commit.sh

jobs-tests: resolve
	go test ./jobs/...

ci-pr: build-image push-image
ci-release: build-image push-image

build-image:
	docker build -t $(IMG_NAME) .
push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)
