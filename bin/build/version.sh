#!/usr/bin/env bash

bare_ref="${GITHUB_REF##*/}"
>&2 echo "Persisting version <$bare_ref>"
if [[ $bare_ref =~ ^v([0-9]+).([0-9]+).([0-9]+)$ ]]; then
  # For a stable tag, grab from github and exit
  echo -n "$bare_ref"
  exit
elif [[ "$bare_ref" == "master" ]]; then
  # from the mainline branch, grab the sha
  sha="$GITHUB_SHA"
else
  # from a local build
  sha="$(git rev-parse HEAD)+local.$(date -u +"%Y%m%dT%H%M%SZ")"
fi

parentTag="$(git describe --abbrev=0 --tags --exclude '*-*' 2>/dev/null || echo -n "v0.0.0")"

# follows https://semver.org/
echo -n "${parentTag}-${sha}"
