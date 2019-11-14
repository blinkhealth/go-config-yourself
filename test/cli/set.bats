#!/usr/bin/env bats
load "conftest"

function testDefaultFile() {
  local variation; variation="$1"

  src="$(fixture encrypted.kms)"
  tmpDir="$(dirname "$src")"
  file="$tmpDir/updates-defaults-file.yml"
  defaultFile="$tmpDir/$variation"

  cp "$src" "$file"
  touch "$defaultFile"
  bc set "$file" newProp <<<"this is very secret"
  bc get $defaultFile newProp
  rm "$defaultFile"
}

@test "set stores plaintext values" {
  file=$(fixture encrypted.kms)
  bc set --plain-text $file myPlainTextString <<<'plaintext'
  grep "plaintext" $file >/dev/null

  # Produces good json
  [[ "$(bc get $file crypto)" == *'{"key":"arn:aws:kms:'* ]]
}

@test "set interprets input as json" {
  file=$(fixture encrypted.kms)
  # extra weird spacing just to see we get a proper yaml array inserted
  bc set --plain-text $file "myPlainTextArray" <<<'[1, 2,3]'
  [ "$status" -eq 0 ]

  bc get $file myPlainTextArray
  [ "$status" -eq 0 ]
  # Now we test to get that same array back, json-encoded
  [[ "$output" == *'[1,2,3]'* ]]
}

@test "set stores encrypted values" {
  file=$(fixture encrypted.kms)
  secret="this is very secret"

  bc set $file myEncryptedKey <<<"$secret"
  [[ "$(bc get $file myEncryptedKey)" == "$secret" ]]
}

@test "set reads files as values" {
  file=$(fixture encrypted.kms)

  bc set --input-file <(echo "asdf") $file encryptedFile
  [[ "$(bc get $file encryptedFile)" == "asdf" ]]
}

@test "set creates a hash for encrypted values" {
  file=$(fixture encrypted.kms)
  secret="this is very secret"
  expectedHash=$(echo -n "$secret" | openssl dgst -sha256 | awk '{print $2}')

  bc set $file hashProp <<<"$secret"
  grep -E "hash:\s*$expectedHash" $file > /dev/null
}

@test "set reads up to 4k out of stdin" {
  file=$(fixture encrypted.kms)
  secret="$(printf '.%.0s' {1..4095})"

  bc set $file longBlob <<<"$secret"

  [[ "$(bc get $file longBlob)" == "$secret" ]]
}

@test "set refuses to update crypto property" {
  file=$(fixture encrypted.kms)

  run $CMD set $file crypto <<<"must-fail"
  [[ $status != 0 ]];

  run $CMD set $file crypto.key <<<"must-fail"
  [[ $status != 0 ]];
}

@test "set updates default file" {
  testDefaultFile default.yml
}

@test "set updates defaults file" {
  testDefaultFile defaults.yml
}

