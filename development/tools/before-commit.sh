#!/usr/bin/env bash

readonly CI_FLAG=ci

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
NC='\033[0m' # No Color
CI_ENABLED=0
if [ "$1" == "$CI_FLAG" ]; then
  CI_ENABLED=1
  shift
fi
DIRS_TO_CHECK=("$@")

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
go test -count=1 "${DIRS_TO_CHECK[@]}"

check_result "go test" $?

##
#  GO LINT
##
go get golang.org/x/lint/golint
buildLintResult=$?
if [ ${buildLintResult} != 0 ]; then
  echo -e "${RED}✗ go get golint${NC}\n$buildLintResult${NC}"
  exit 1
fi

golintResult=$("${GOPATH}"/bin/golint "${DIRS_TO_CHECK[@]}")

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

dirs=$(go list -f '{{ .Dir }}' "${DIRS_TO_CHECK[@]}" | grep -E -v "/vendor|/automock|/testdata")
goimportsCmd="$(for d in $dirs; do "${GOPATH}"/goimports -l "$d"/*.go; done)"
goImportsResult=$(test -z "$goimportsCmd") # check if result of command is empty

if [ "$goImportsResult" != 0 ]; then
  echo -e "${RED}✗ goimports and fmt ${NC}\n$goImportsResult${NC}"
    echo -e "changed files: \n$goimportsCmd"
    echo "run goimports against the development/tools/ and commit your changes"
  exit 1
else
  echo -e "${GREEN}√ goimports and fmt ${NC}"
fi

##
# GO VET
##
echo "? go vet"
for vPackage in "${DIRS_TO_CHECK[@]}"; do
  vetResult=$(go vet "${vPackage}")
  check_result "go vet ${vPackage}" "${#vetResult}" "${vetResult}"
done

echo "? go build"
buildEnv=""
if [ $CI_ENABLED -eq 1 ]; then
  # build binary statically
  buildEnv="env CGO_ENABLED=0"
fi

while IFS= read -r -d '' directory; do
  cmdName=$(basename "${directory}")
  if [ -a "${directory}/nobuild.lock" ]; then
    continue
  fi
  ${buildEnv} go build -o "${cmdName}" "${directory}"
  buildResult=$?
  rm "${cmdName}"
  check_result "go build ${directory}" "${buildResult}"
done < <(find "./cmd" -mindepth 1 -type d -print0)