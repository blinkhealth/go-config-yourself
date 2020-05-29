#!/usr/bin/env bats
load "conftest"

@test "get reads plaintext values" {
  file=$(fixture encrypted.kms)
  bc set --plain-text $file myPlainTextString <<<'a plaintext string'
  grep "myPlainTextString: a plaintext string" $file >/dev/null
}

@test "get outputs valid json" {
  file=$(fixture encrypted.kms)
  echo "$(bc get $file crypto)"
  [[ "$(bc get $file crypto)" == *'{"key":"arn:aws:kms:'* ]]
}

@test "get reads encrypted values" {
  file=$(fixture encrypted.kms)
  bc set $file newSecret <<<"a new secret"
  [[ "$(bc get $file newSecret)" == *'a new secret'* ]]
}

@test "get literal dot reads everything" {
  file=$(fixture encrypted.kms)
  [[ "$(bc get $file . | jq -r 'keys | length')" == '9' ]]
  [[ "$(bc get $file . | openssl dgst -md5)" == '687d38244ad3eb2fd96cd93f4dafd012' ]]
}
