.PHONY: validate
validate:
	./before-commit.sh ci

.PHONY: ci-pr
ci-pr: validate

.PHONY: ci-main
ci-main: validate

.PHONY: ci-release
ci-release: validate
