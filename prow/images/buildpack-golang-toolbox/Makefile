IMG_NAME = buildpack-golang-toolbox
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

ci-pr: build-image push-image
ci-release: build-image push-image-release

build-image:
	docker build -t $(IMG_NAME) .
push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)
push-image-release: push-image
	docker tag $(IMG_NAME) $(IMG):go1.12-v1
	docker push $(IMG):go1.12-v1