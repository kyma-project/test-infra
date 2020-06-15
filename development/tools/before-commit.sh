#!/usr/bin/env bash

readonly CI_FLAG=ci

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

echo -e "${INVERTED}"
echo "USER: ${USER}"
echo "PATH: ${PATH}"
echo "GOPATH: ${GOPATH}"
echo -e "${NC}"

function check_result() {
    local step=$1
    local result=$2
    local output=$3

    if [[ ${result} != 0 ]]; then
        echo -e "${RED}✗ ${step}${NC}\\n${output}"
        exit 1
    else
        echo -e "${GREEN}√ ${step}${NC}"
    fi
}

goFilesToCheck=$(find . -type f -name "*.go" | grep -E -v "/vendor/|/automock/|/testdata/")

##
# GO MOD VERIFY
##
go mod verify
ensureResult=$?
if [ ${ensureResult} != 0 ]; then
  echo -e "${RED}✗ go mod verify${NC}\n$ensureResult${NC}"
  exit 1
else
  echo -e "${GREEN}√ go mod verify${NC}"
fi

##
# GO BUILD
##

##
# GO TEST
##
echo "? go test"
if [ "$1" == "$CI_FLAG" ]; then
  go test -count=1 ./pkg/...
else
  go test -count=1 ./...
fi
check_result "go test" $?

echo "? go build"
buildEnv=""
if [ "$1" == "$CI_FLAG" ]; then
	# build binary statically
	buildEnv="env CGO_ENABLED=0"
fi

while IFS= read -r -d '' directory
do
    cmdName=$(basename "${directory}")
    if [ -a "${directory}/nobuild.lock" ]; then
      continue
    fi
    ${buildEnv} go build -o "${cmdName}" "${directory}"
    buildResult=$?
    rm "${cmdName}"
    check_result "go build ${directory}" "${buildResult}"
done <   <(find "./cmd" -mindepth 1 -type d -print0)

##
#  GO LINT
##
go get golang.org/x/lint/golint
buildLintResult=$?
if [ ${buildLintResult} != 0 ]; then
  echo -e "${RED}✗ go get golint${NC}\n$buildLintResult${NC}"
  exit 1
fi

golintResult=$(echo "${goFilesToCheck}" | xargs -L1 "${GOPATH}"/bin/golint)

if [ "${#golintResult}" != 0 ]; then
  echo -e "${RED}✗ golint\n$golintResult${NC}"
else
  echo -e "${GREEN}√ golint${NC}"
fi

##
# GO IMPORTS & FMT
##
go get golang.org/x/tools/cmd/goimports
buildGoImportResult=$?
if [ ${buildGoImportResult} != 0 ]; then
  echo -e "${RED}✗ go build goimports${NC}\n$buildGoImportResult${NC}"
  exit 1
fi

goImportsResult=$(echo "${goFilesToCheck}" | xargs -L1 "${GOPATH}"/bin/goimports -w -l)

if [ "${#goImportsResult}" != 0 ]; then
  echo -e "${RED}✗ goimports and fmt ${NC}\n$goImportsResult${NC}"
  exit 1
else
  echo -e "${GREEN}√ goimports and fmt ${NC}"
fi
##
# GO VET
##
echo "? go vet"
packagesToVet=("./cmd/..." "./jobs/..." "./pkg/...")

for vPackage in "${packagesToVet[@]}"; do
	vetResult=$(go vet "${vPackage}")
    check_result "go vet ${vPackage}" "${#vetResult}" "${vetResult}"
done

