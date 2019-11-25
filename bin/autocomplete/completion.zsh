#compdef gcy
#autoload
# shellcheck shell=bash

_gcy_zsh_autocomplete () {

  local -a opts
  opts=("${(@f)$(_CLI_ZSH_AUTOCOMPLETE_HACK=1 ${words[@]:0:#words[@]-1} --generate-bash-completion)}")

  exit_code="$?"
  if [[ $exit_code -gt 0 ]]; then
    _path_files
    [[ $exit_code == 1 ]]; return
  fi

  _describe 'gcy' opts

  return
}

compdef _gcy_zsh_autocomplete gcy
