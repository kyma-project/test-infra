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
binaries=("logs-printer" "stability-checker")
buildEnv=""
if [ "$1" == "$CI_FLAG" ]; then
  # build binary statically
  buildEnv="env CGO_ENABLED=0"
fi

for binary in "${binaries[@]}"; do
  ${buildEnv} go build -o "${binary}" ./cmd/"${binary}"
  goBuildResult=$?
  if [ ${goBuildResult} != 0 ]; then
    echo -e "${RED}✗ go build ${binary} ${NC}\n $goBuildResult${NC}"
    exit 1
  else
    echo -e "${GREEN}√ go build ${binary} ${NC}"
  fi
done

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
  exit 1
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
packagesToVet=("./cmd/..." "./internal/...")

for vPackage in "${packagesToVet[@]}"; do
  vetResult=$(go vet "${vPackage}")
  if [ "${#vetResult}" != 0 ]; then
    echo -e "${RED}✗ go vet ${vPackage} ${NC}\n$vetResult${NC}"
    exit 1
  else
    echo -e "${GREEN}√ go vet ${vPackage} ${NC}"
  fi
done

##
# GO TEST
##
echo "? go test"
go test ./...
# Check if tests passed
if [ $? != 0 ]; then
  echo -e "${RED}✗ go test\n${NC}"
  exit 1
else
  echo -e "${GREEN}√ go test${NC}"
fi

goFilesToCheck=$(find . -type f -name "*.go" | egrep -v "\/vendor\/|_*/automock/|_*/testdata/|/pkg\/|_*export_test.go")
