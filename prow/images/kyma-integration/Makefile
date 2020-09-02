# Environment variables expected by this makefile
# DOCKER_TAG - dynamic tag for image
# DOCKER_PUSH_REPOSITORY - URL set by prow preset
# DOCKER_PUSH_DIRECTORY

IMG_NAME = kyma-integration
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)

# build and push on PR, tag with PR number
.PHONY: ci-pr
ci-pr: build-image push-image
push-image : build-image


# build and push on postsubmit, tag with commit ID and current
.PHONY: ci-release
ci-release: build-image tag-image push-image
tag-image: build-image


# tag image with current tag
.PHONY: tag-image
tag-image:
	docker tag $(IMG):$(DOCKER_TAG)-k8s1.16 $(IMG):current
	docker tag $(IMG):$(DOCKER_TAG)-k8s1.16 $(IMG):k8s1.16-current
	docker tag $(IMG):$(DOCKER_TAG)-k8s1.15 $(IMG):k8s1.15-current
	docker tag $(IMG):$(DOCKER_TAG)-k8s1.14 $(IMG):k8s1.14-current


# build image and tag it with commit ID or PR number
.PHONY: build-image
build-image:
	docker build -t $(IMG):$(DOCKER_TAG)-k8s1.16 --build-arg K8S_VERSION=1.16 --build-arg commit=$(DOCKER_TAG) .
	docker build -t $(IMG):$(DOCKER_TAG)-k8s1.15 --build-arg K8S_VERSION=1.15 --build-arg commit=$(DOCKER_TAG) .
	docker build -t $(IMG):$(DOCKER_TAG)-k8s1.14 --build-arg K8S_VERSION=1.14 --build-arg commit=$(DOCKER_TAG) .


# push image with all tags
.PHONY:push-image
push-image:
	docker push $(IMG)
