#!/usr/bin/env bash

function github::configure_git() {
  echo "Configuring git"
  # configure ssh
  if [[ -n "${BOT_GITHUB_SSH_PATH}" ]]; then
      mkdir "${HOME}/.ssh/"
      cp "${BOT_GITHUB_SSH_PATH}" "${HOME}/.ssh/ssh_key.pem"
      local SSH_FILE="${HOME}/.ssh/ssh_key.pem"
      touch "${HOME}/.ssh/known_hosts"
      ssh-keyscan -H github.com >> "${HOME}/.ssh/known_hosts"
      chmod 400 "${SSH_FILE}"
      eval "$(ssh-agent -s)"
      ssh-add "${SSH_FILE}"
      ssh-add -l
      git config --global core.sshCommand "ssh -i ${SSH_FILE}"
  fi

  # configure email
  if [[ -n "${BOT_GITHUB_EMAIL}" ]]; then
      git config --global user.email "${BOT_GITHUB_EMAIL}"
  fi

  # configure name
  if [[ -n "${BOT_GITHUB_NAME}" ]]; then
      git config --global user.name "${BOT_GITHUB_NAME}"
  fi
}
