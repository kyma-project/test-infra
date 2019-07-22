DOCKER_PUSH_REPOSITORY = eu.gcr.io/kyma-project/test-infra
IMG_NAME = fluentd-gcp
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(IMG_NAME)
TAG = v20190719

ci-pr: build-image push-image
ci-release: build-image push-image

 build-image:
	docker build -t $(IMG_NAME) .
push-image:
	docker tag $(IMG_NAME) $(IMG):$(TAG)
	docker push $(IMG):$(TAG)  
