IMG_NAME = kyma-integration
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = $(DOCKER_TAG)

.PHONY: ci-pr
ci-pr: build-image push-image
push-image : build-image


.PHONY: ci-release
ci-release: build-image tag-image push-image
tag-image: build-image
push-image: tag-image


.PHONY: tag-image
tag-image:
	docker tag $(IMG):$(TAG)-k8s1.16 $(IMG):current
	docker tag $(IMG):$(TAG)-k8s1.16 $(IMG):k8s1.16-current
	docker tag $(IMG):$(TAG)-k8s1.15 $(IMG):k8s1.15-current
	docker tag $(IMG):$(TAG)-k8s1.14 $(IMG):k8s1.14-current


.PHONY: build-image
build-image:
	docker build \
		--cache-from $(IMG):latest\
		--build-arg K8S_VERSION=1.16 \
		--build-arg commit=$(TAG) \
		--tag $(IMG):$(TAG)-k8s1.16 \
		--tag $(IMG):latest \
		.
	docker build \
		--cache-from $(IMG):latest\
		--build-arg K8S_VERSION=1.15 \
		--build-arg commit=$(TAG) \
		--tag $(IMG):$(TAG)-k8s1.15 \
		--tag $(IMG):latest \
		.
	docker build \
		--cache-from $(IMG):latest\
		--build-arg K8S_VERSION=1.14 \
		--build-arg commit=$(TAG) \
		--tag $(IMG):$(TAG)-k8s1.14 \
		--tag $(IMG):latest \
		.


.PHONY:push-image
push-image:
	docker push $(IMG) # it pushes all tags for defined image
