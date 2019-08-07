#!/usr/bin/env bats
TESTCMD="go run -tags test main.go"
export CMD=${INVOKE_CMD:-$TESTCMD}
export WORKDIR="/tmp/go-config-yourself-test"
# For the source of these values, look at provider/kms/service_kms_mock.go
export GOOD_KEY="arn:aws:kms:us-east-1:000000000000:key/00000000-0000-0000-0000-000000000000"
export BAD_KEY="arn:aws:kms:us-east-1:111111111111:key/00000000-0000-0000-0000-111111111111"
mkdir -p "$WORKDIR"

function setup() {
  # use mock creds
  export AWS_ACCESS_KEY_ID="AGOODACCESSKEYID"
  export AWS_SECRET_ACCESS_KEY="AVERYSECRETACCESSKEYTHATSNOTBASE64ENC="
  export AWS_SHARED_CREDENTIALS_FILE="/dev/null"
  unset AWS_SESSION_TOKEN
  unset AWS_PROFILE
}

function fixture() {
  local name dest; name="$1"
  dest="${WORKDIR}/$(cd "$WORKDIR" && mktemp "go-config-yourself-test.XXXXXX")"
  cp "test/fixtures/$name.yaml" "$dest"
  echo "$dest"
}

function teardown() {
  rm -rf ${WORKDIR:?}/*
}

function bc () {
  run $CMD ${*}

  if [[ "$status" -ne 0 ]]; then
    echo "----"
    echo "error: $CMD ${*}"
    echo "status: $status"
    printf '%s\n' "${lines[@]}"
    echo "----"
    return "$status"
  fi

  for line in "${lines[@]}"; do
    if [[ $line != 'level'* ]]; then
      printf '%s\n' "$line"
    fi
  done
}
