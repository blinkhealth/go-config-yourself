#compdef gcy
#autoload
# shellcheck shell=bash

_gcy_zsh_autocomplete () {
  local -a opts
  cur="${words[-1]}"
  opts=("${(@f)$(_CLI_ZSH_AUTOCOMPLETE_HACK=1 CUR="$cur" ${words[@]:0:#words[@]-1} --generate-bash-completion)}")

  exit_code="$?"
  if [[ $exit_code -gt 0 ]]; then
    _path_files
    return
  fi

  _describe "qwer" opts
}

compdef _gcy_zsh_autocomplete gcy
