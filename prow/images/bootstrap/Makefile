IMG_NAME = bootstrap
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

# biuld and push on PR, tag with PR number
.PHONY: ci-pr
ci-pr: build-image push-image
push-image: build-image


# build and push on postsubmit, tag with commit ID and current
.PHONY: ci-release
ci-release: build-image tag-image push-image
tag-image: build-image
push-image: tag-image


# tag image with current tag
.PHONY: tag-image
tag-image:
	docker tag $(IMG):$(TAG) $(IMG):current


# build image and tag it with commit ID or PR number
.PHONY: build-image
build-image:
	docker build -t $(IMG):$(TAG) --build-arg commit=$(TAG) .


# push image with all tags
.PHONY: push-image
push-image:
	docker push $(IMG)