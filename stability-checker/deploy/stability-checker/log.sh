#!/usr/bin/env bash

# COPIED FROM test-infra/prow/scripts/lib/log.sh

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
YELLOW='\e[33m'
NC='\033[0m' # No Color

# log::date retruns current date in format expected by logs
function log::date {
    date +"%Y/%m/%d %T %Z"
}

# log::info prints message with info level
#
# Arguments:
#   $1 - Message
function log::info {
    echo -e "${INVERTED}$(log::date) [INFO] ${1}${NC}"
}

# log::info prints a message with info level in green
#
# Arguments:
#   $1 - Message
function log::success {
    echo -e "${GREEN}$(log::date) [INFO] ${1}${NC}"
}

# log::info prints a message with warning level in yellow
#
# Arguments:
#   $1 - Message
function log::warn {
    echo -e "${YELLOW}$(log::date) [WARN] ${1}${NC}"
}

# log::info prints a message with error level in red
#
# Arguments:
#   $1 - Message
function log::error {
    >&2 echo -e "${RED}$(log::date) [ERRO] ${1}${NC}"
}
