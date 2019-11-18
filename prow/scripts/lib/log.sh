#!/usr/bin/env bash

readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly INVERTED='\033[7m'
readonly YELLOW='\e[33m'
readonly NC='\033[0m' # No Color

function log::date {
    date +"%Y/%m/%d %T %Z"
}

function log::info {
    echo -e "${INVERTED}$(log::date) [INFO] ${1}${NC}"
}

function log::success {
    echo -e "${GREEN}$(log::date) [INFO] ${1}${NC}"
}


function log::warn {
    echo -e "${YELLOW}$(log::date) [WARN] ${1}${NC}"
}

function log::error {
    >&2 echo -e "${RED}$(log::date) [ERRO] ${1}${NC}"
}
