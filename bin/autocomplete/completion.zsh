#compdef gcy
# shellcheck shell=bash
autoload -U compinit && compinit;
autoload -U bashcompinit && bashcompinit;

_cli_bash_autocomplete() {
  local cur opts;
  COMPREPLY=();
  cur="${COMP_WORDS[COMP_CWORD]}";
  opts=$( CUR="$cur" "${COMP_WORDS[@]:0:$COMP_CWORD}" --generate-completion );
  COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) );
  return 0;
};

_gcy () {
  complete -o nospace -o default -F _cli_bash_autocomplete gcy
}
