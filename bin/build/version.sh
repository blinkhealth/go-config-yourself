#!/usr/bin/env bash

if [[ $CIRCLE_TAG != "" ]]; then
  # when we're building a stable release
  echo -n "$CIRCLE_TAG"
  exit
fi

if [[ "$CIRCLE_SHA1" != "" ]]; then
  # we get here on circleci when building from the mainline branch
  sha="$CIRCLE_SHA1"
else
  # we get here building locally, i.e. make build-local
  sha="$(git rev-parse HEAD)+local.$(date -u +"%Y%m%dT%H%M%SZ")"
fi

parentTag="$(git describe --abbrev=0 --tags --exclude '*-*' 2>/dev/null || echo -n "v0.0.0")"

# follows https://semver.org/
echo -n "${parentTag}-${sha}"
