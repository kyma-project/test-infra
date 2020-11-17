#!/usr/bin/env bash

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
  echo -e "${logdate} [INFO] *************************************************************************************"
  echo -e "${logdate} [INFO] * $*"
  echo -e "${logdate} [INFO] *************************************************************************************"
}

# log::info prints message with info level
#
# Arguments:
#   $* - Message
function log::info {
    echo -e "$(log::date) [INFO] $*"
}

# log::success prints a message with info level
#
# Arguments:
#   $* - Message
function log::success {
  local logdate
  logdate=$(log::date)
  echo -e "${logdate} [INFO] ====================================================================================="
  echo -e "${logdate} [INFO] = SUCCESS                                                                           ="
  echo -e "${logdate} [INFO] ====================================================================================="
  echo -e "${logdate} [INFO] = $*"
  echo -e "${logdate} [INFO] ====================================================================================="
}

# log::warn prints a message with warning level
#
# Arguments:
#   $* - Message
function log::warn {
    echo -e "$(log::date) [WARN] $*"
}

# log::error prints a message with error level
#
# Arguments:
#   $* - Message
function log::error {
    >&2 echo -e "$(log::date) [ERROR] $*"
}
