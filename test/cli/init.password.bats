#!/usr/bin/env bats
load "conftest"

@test "password: init creates a valid config file with strong passwords" {
  file=$(fixture encrypted.password)
  rm "$file"

  # flag
  bc init --password '$GOOD_PASSWORD' --provider password $file
  rm "$file"
  # env var
  CONFIG_PASSWORD="$GOOD_PASSWORD" bc init --provider password $file
}

@test "password: init fails with insecure passwords" {
  file=$(fixture encrypted.password)
  rm "$file"

  run $CMD init --password "vryshrt" --provider password $file
  echo $output
  [[ $status != 0 ]];
  [[ $output == *"Chosen password is too short"* ]]

  run $CMD init --password "aaaabbbbbccccc" --provider password $file
  echo $output
  [[ $status != 0 ]];
  [[ $output == *"Password seems easy to guess"* ]]

  run $CMD init --password "electrodoméstico" --provider password $file
  echo $output
  [[ $status != 0 ]];
  [[ $output == *"Password is too common"* ]]
}


@test "password: init respects --skip-password-validation" {
  file=$(fixture encrypted.password)
  rm "$file"

  bc init --skip-password-validation --password "password" --provider password $file
  rm "$file"
}
