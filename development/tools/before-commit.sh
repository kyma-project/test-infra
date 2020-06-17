#!/usr/bin/env bash

# before-commit.sh executes basic tests and code format checking against go files in provided directory.
# It accepts go directory syntax. For example to check all files under ./pkg/ directory use argument "./pkg/..."
#
# Usage: ./before-commit.sh ./dir/... ./dir/...

readonly CI_FLAG=ci

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color
if [ "$#" -eq 0 ];
then
  echo "No arguments passed! Continuing with all directories in current directory!"
  DIRS_TO_CHECK=("./...")
else
  DIRS_TO_CHECK=("$@")
fi

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

##
# GO MOD VERIFY
##
echo "? go mod verify"
go mod verify
ensureResult=$?
if [ ${ensureResult} != 0 ]; then
  echo -e "${RED}✗ go mod verify${NC}\n$ensureResult${NC}"
  exit 1
else
  echo -e "${GREEN}√ go mod verify${NC}"
fi

##
# GO TEST
##
echo "? go test"
go test -count=1 "${DIRS_TO_CHECK[@]}"

check_result "go test" $?

##
#  GO LINT
##
echo "? golint"
go get golang.org/x/lint/golint
buildLintResult=$?
if [ ${buildLintResult} != 0 ]; then
  echo -e "${RED}✗ go get golint${NC}\n$buildLintResult${NC}"
  exit 1
fi

golintResult=$("${GOPATH}"/bin/golint "${DIRS_TO_CHECK[@]}")
if [ "${#golintResult}" != 0 ]; then
  echo -e "${RED}✗ golint\n$golintResult${NC}"
  exit 1
else
  echo -e "${GREEN}√ golint${NC}"
fi

##
# GO IMPORTS & FMT
##
echo "? goimports"
go get golang.org/x/tools/cmd/goimports
buildGoImportResult=$?
if [ ${buildGoImportResult} != 0 ]; then
  echo -e "${RED}✗ go build goimports${NC}\n$buildGoImportResult${NC}"
  exit 1
fi

dirs=$(go list -f '{{ .Dir }}' "${DIRS_TO_CHECK[@]}" | grep -E -v "/vendor|/automock|/testdata")
changedFiles=$(for d in $dirs; do "${GOPATH}"/bin/goimports -l "$d"/*.go; done)
test -z "$changedFiles"
goImportsResult=$?

if [ "$goImportsResult" != 0 ]; then
  echo -e "${RED}✗ goimports ${NC}\n$goImportsResult${NC}"
    echo "changed files:"
    echo "$changedFiles"
    echo "run \"goimports -w -l\" against the development/tools/ and commit your changes"
  exit 1
else
  echo -e "${GREEN}√ goimports${NC}"
fi

##
# GO VET
##
echo "? go vet"
for vPackage in "${DIRS_TO_CHECK[@]}"; do
  vetResult=$(go vet "${vPackage}")
  check_result "go vet ${vPackage}" "${#vetResult}" "${vetResult}"
done
