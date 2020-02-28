#!/usr/bin/env bash

readonly CI_FLAG=ci

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"


RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

echo -e "${INVERTED}"
echo "USER: " + $USER
echo "PATH: " + $PATH
echo "GOPATH:" + $GOPATH
echo -e "${NC}"

cd ${DIR}

##
# Tidy dependencies
##
go mod tidy
ensureResult=$?
if [ ${ensureResult} != 0 ]; then
	echo -e "${RED}✗ go mod tidy${NC}\n$ensureResult${NC}"
	exit 1
else echo -e "${GREEN}√ go mod tidy${NC}"
fi

## Ensure go mod tidy did not modify go.mod and go.sum
if [[ "$1" == "$CI_FLAG" ]]; then
  if [[ -n $(git status -s go.*) ]]; then
		echo -e "${RED}✗ go mod tidy modified go.mod or go.sum files${NC}";
    exit 1
  fi
fi


##
# Validate dependencies
##
echo "? go mod verify"
depResult=$(go mod verify)
if [ $? != 0 ]; then
	echo -e "${RED}✗ go mod verify\n$depResult${NC}"
	exit 1
else echo -e "${GREEN}√ go mod verify${NC}"
fi

##
# GO TEST
##
echo "? go test"
go test -race -coverprofile=cover.out ./...
# Check if tests passed
if [[ $? != 0 ]]; then
	echo -e "${RED}✗ go test\n${NC}"
	rm cover.out
	exit 1
else
	echo -e "Total coverage: $(go tool cover -func=cover.out | grep total | awk '{print $3}')"
	rm cover.out
	echo -e "${GREEN}√ go test${NC}"
fi

goFilesToCheck=$(find . -type f -name "*.go" | egrep -v "\/vendor\/|_*/automock/|_*/testdata/|_*export_test.go|mock_api.go")

#
# GO FMT
#
goFmtResult=$(echo "${goFilesToCheck}" | xargs -L1 go fmt)
if [ "${#vetResult}" != 0 ]
	then
    	echo -e "${RED}✗ go fmt${NC}\n$goFmtResult${NC}"
    	exit 1;
	else echo -e "${GREEN}√ go fmt${NC}"
fi

##
#  GO LINT
##
echo "? golint"
golintResult=$(echo "${goFilesToCheck}" | xargs -L1 golint)
if [ "${#golintResult}" != 0 ]; then
    echo -e "${RED}✗ golint${NC}\\n${golintResult}"
else
    echo -e "${GREEN}√ golint${NC}"
fi

##
# GO VET
##
packagesToVet=("./cmd/..." "./pkg/...")

for vPackage in "${packagesToVet[@]}"; do
	vetResult=$(go vet "${vPackage}")
	if [ "${#vetResult}" != 0 ]; then
		echo -e "${RED}✗ go vet ${vPackage} ${NC}\n$vetResult${NC}"
		exit 1
	else echo -e "${GREEN}√ go vet ${vPackage} ${NC}"
	fi

done
