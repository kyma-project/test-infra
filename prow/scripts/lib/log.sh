#!/usr/bin/env bash

# log::date returns the current date in format expected by logs
function log::date {
    date +"%Y-%m-%d %T %Z"
}

# log::dump_trace prints stacktrace when an error occurs.
#
log::dump_trace() {
    local frame=1 line func source n=0
    while caller "$frame"; do
        ((frame++))
    done | while read -r line func source; do
        ((n++ == 0)) && {
            printf 'Encountered an error\n'
        }
        printf '%4s at %s\n' " " "$func ($source:$line)"
    done
}

# log::banner prints message with INFO level in banner form for easier spotting in log files
#
# Arguments:
#   $* - Message
function log::banner {
  local logdate
  logdate=$(log::date)
  local scriptname
  scriptname=${BASH_SOURCE[1]:-$1}
  echo -e "${logdate} [INFO] *************************************************************************************"
  echo -e "${logdate} [INFO] [$scriptname] * $*"
  echo -e "${logdate} [INFO] *************************************************************************************"
}

# log::info prints message with info level
#
# Arguments:
#   $* - Message
function log::info {
   local funcname # get function that called this
   local scriptname
   funcname=${FUNCNAME[1]}
   scriptname=${BASH_SOURCE[1]:-$1}
   echo -e "$(log::date) [INFO] PID:$$ --- [$scriptname] $funcname:${BASH_LINENO[1]} $*"
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
   local funcname # get function that called this
   local scriptname
   funcname=${FUNCNAME[1]}
   scriptname=${BASH_SOURCE[1]:-$1}
  echo -e "$(log::date) [WARN] PID:$$ --- [$scriptname] $funcname:${BASH_LINENO[1]} $*"
}

# log::error prints a message with error level
#
# Arguments:
#   $* - Message
function log::error {
     local funcname # get function that called this
     local scriptname
     funcname=${FUNCNAME[1]}
     scriptname=${BASH_SOURCE[1]:-$1}
    >&2  echo -e "$(log::date) [ERROR] PID:$$ --- [$scriptname] $funcname:${BASH_LINENO[1]} $*"
    >&2 log::dump_trace
}# (2025-03-04)