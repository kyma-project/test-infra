#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
INVERTED='\033[7m'
YELLOW='\e[33m'
NC='\033[0m' # No Color

# log::date retruns current date in format expected by logs
function log::date {
    date +"%Y/%m/%d %T %Z"
}

# log::banner prints message with INFO level in banner form for easier spotting in log files
#
# Arguments:
#   $* - Message
function log::banner {
  local logdate
  logdate=$(log::date)
  echo -e "${INVERTED}${logdate} [INFO] *************************************************************************************${NC}"
  echo -e "${INVERTED}${logdate} [INFO] * $* ${NC}"
  echo -e "${INVERTED}${logdate} [INFO] *************************************************************************************${NC}"
}

# log::info prints message with info level
#
# Arguments:
#   $* - Message
function log::info {
    echo -e "${INVERTED}$(log::date) [INFO] $* ${NC}"
}

# log::info prints a message with info level in green
#
# Arguments:
#   $* - Message
function log::success {
    echo -e "${GREEN}$(log::date) [INFO] $* ${NC}"
}

# log::info prints a message with warning level in yellow
#
# Arguments:
#   $* - Message
function log::warn {
    echo -e "${YELLOW}$(log::date) [WARN] $* ${NC}"
}

# log::info prints a message with error level in red
#
# Arguments:
#   $* - Message
function log::error {
    >&2 echo -e "${RED}$(log::date) [ERRO] $* ${NC}"
}
