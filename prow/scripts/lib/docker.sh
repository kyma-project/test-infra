#!/usr/bin/env bash

# docker::start starts the Docker Daemon if not already started
function docker::start {
    if docker info > /dev/null 2>&1 ; then
        echo "Docker already started"
        return 0
    fi

    echo "Initializing Docker..."
    printf '=%.0s' {1..80}; echo
    # If we have opted in to docker in docker, start the docker daemon,
    service docker start
    # the service can be started but the docker socket not ready, wait for ready
    local WAIT_N=0
    local MAX_WAIT=20
    while true; do
        # docker ps -q should only work if the daemon is ready
        docker ps -q > /dev/null 2>&1 && break
        if [[ ${WAIT_N} -lt ${MAX_WAIT} ]]; then
            WAIT_N=$((WAIT_N+1))
            echo "Waiting for docker to be ready, sleeping for ${WAIT_N} seconds."
            sleep ${WAIT_N}
        else
            echo "Reached maximum attempts, not waiting any longer..."
            return 1
        fi
    done
    printf '=%.0s' {1..80}; echo

    docker-credential-gcr configure-docker
    echo "Done starting up docker."
}

# docker::print_processes prints running docker containers
function docker::print_processes {
    docker ps -a
}

# docker::build_post_pr_tag builds pr tag on postsubmit jobs
function docker::build_post_pr_tag {
  set -e
  log::info "Checking if prtagbuilder binary is present"
  if [ -x /prow-tools/prtagbuilder ]; then
    log::info "Binary prtagbuilder found. Building PR tag."
    DOCKER_POST_PR_TAG=$(/prow-tools/prtagbuilder)
    prTagBuilt=$?
    if [[ $prTagBuilt -eq 0 ]]; then
      readonly DOCKER_POST_PR_TAG
      export DOCKER_POST_PR_TAG
      log::success "PR tag exported as DOCKER_POST_PR_TAG variable."
      set +e
      return 0
    else
      log:error "Failed building PR tag."
      set +e
      return 1
    fi
  else
    log::info "Binary prtagbuilder not found. Trying run prtagbuilder from source."
  fi
  log::info "Checking if go is installed."
  command -v go &> /dev/nul
  goPresent=$?
  if [[ $goPresent -eq 0 ]]; then
    log::info "go installed"
    log::info "Checking if prtagbuilder source file is present"
    BUILDER_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}/../../development/tools/cmd/prtagbuilder" )" && pwd )"
    if [ -a "${BUILDER_DIR}/main.go" ]; then
      log::info "Prtagbuilder source file found. Building PR tag. "
      cd "${BUILDER_DIR}" || exit
      DOCKER_POST_PR_TAG=$(go run main.go)
      prTagBuilt=$?
      if [[ $prTagBuilt -eq 0 ]]; then
        readonly DOCKER_POST_PR_TAG
        export DOCKER_POST_PR_TAG
        log::success "PR tag exported as DOCKER_POST_PR_TAG variable."
        set +e
        return 0
      else
        log:error "Failed building PR tag."
        set +e
        return 1
      fi
    else
      log::error "Prtagbuilder source file not found. Can't run prtagbuilder from source."
      set +e
      return 1
    fi
  else
    log::error "go not installed. Can't run prtagbuilder from source."
    set +e
    return 1
  fi
}