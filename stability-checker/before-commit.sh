#!/usr/bin/env bash

readonly CI_FLAG=ci

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# ensure $GOPATH/bin is present in PATH
export PATH=$GOPATH/bin:$PATH

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color

echo -e "${INVERTED}"
echo "USER: " + "$USER"
echo "PATH: " + "$PATH"
echo "GOPATH:" + "$GOPATH"
echo "CURRENT DIRECTORY: $DIR"
echo -e "${NC}"

cd "${DIR}" || exit 1

##
# GO GENERATE
##
go generate ./...
generateResult=$?
if [ ${generateResult} != 0 ]; then
	echo -e "${RED}✗ go generate ./...${NC}\n$generateResult${NC}"
	exit 1
else echo -e "${GREEN}√ go generate ./...${NC}"
fi

##
# Tidy dependencies
##
echo "? go mod tidy"
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
go mod verify
verifyResult=$?
if [ ${ensureResult} != 0 ]; then
	echo -e "${RED}✗ go mod verify\n$verifyResult${NC}"
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
go get -u golang.org/x/lint/golint
golintResult=$(echo "${goFilesToCheck}" | xargs -L1 golint)
if [ "${#golintResult}" != 0 ]; then
    echo -e "${RED}✗ golint${NC}\\n${golintResult}"
else
    echo -e "${GREEN}√ golint${NC}"
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
	else echo -e "${GREEN}√ go vet ${vPackage} ${NC}"
	fi

done
