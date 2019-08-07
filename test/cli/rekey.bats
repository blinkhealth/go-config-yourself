#!/usr/bin/env bats
load "conftest"

@test "rekey works between providers" {
  file=$(fixture encrypted.kms)
  old_ciphertext=$(grep ciphertext $file)
  export CONFIG_PASSWORD="asdf"
  bc rekey --provider password $file

  # cyphertext changed
  [[ $(grep ciphertext $file) != $old_ciphertext ]]
  # hash didnt
  grep "hash:\s*$HASH" $file >/dev/null
}
