.PHONY: resolve validate jobs-tests

cmds := $(foreach cmd,$(shell go list -f '{{ if eq .Name "main" }}{{ .ImportPath }}{{end}}' ./...),$(notdir $(cmd)))
testpkgs := $(shell go list -test -f '{{ if .TestGoFiles }}{{ .ForTest }}{{end}}' ./... | sed s%github.com/kyma-project/test-infra/development/tools/%% | xargs -n1 dirname | sort -u )

cmds_run = $(foreach cmd,$(cmds),$(cmd)-run)
testpkgs_test = $(foreach cmd,$(testpkgs),$(cmd)-tests)

resolve:
	dep ensure -v -vendor-only

validate: resolve
	./before-commit.sh

$(cmds): resolve
$(cmds): vendor
	@echo "+ Building $@"
	@CGO_ENABLED=0 go build -o $@ \
		./cmd/$@

$(cmds_run): %-run: vendor
	@echo "+ Running $*"
	@CGO_ENABLED=0 go run ./cmd/$* ${ARGS}

$(testpkgs_test): %-tests:
	@echo "+ testing $*"
	go test -v ./$*/...

.PHONY: list
list:
	@$(MAKE) -pRrq -f Makefile : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$'
