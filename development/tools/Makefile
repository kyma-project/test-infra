APP_NAME ?= prow-tools

.PHONY: validate
validate:
	./before-commit.sh ./cmd/... ./pkg/...

.PHONY: clean
clean:
	rm -f $(APP_NAME)

