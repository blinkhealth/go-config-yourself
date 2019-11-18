#!/usr/bin/env bats
load "conftest"

@test "kms: init fails when no aws credentials are set" {
  export AWS_ACCESS_KEY_ID=""
  export AWS_PROFILE=""
  file=$(fixture encrypted.kms)
  rm "$file"

  [[ "$(bc init $file)" == *"No AWS credentials found"* ]]
}

@test "kms: init fails when no destination is given" {
  [[ "$(bc init)" == *"Destination to save config file missing"* ]]
}

@test "kms: init creates a valid config file without a provider" {
  file=$(fixture encrypted.kms)
  rm "$file"

  bc init --key $GOOD_KEY $file
  grep us-east-1 $file
}

@test "kms: init creates a valid config file with a provider" {
  file=$(fixture encrypted.kms)
  rm "$file"

  bc init --key $GOOD_KEY --provider kms $file
  grep us-east-1 $file
}

@test "kms: init creates a valid config file with arguments" {
  file=$(fixture encrypted.kms)
  rm "$file"

  bc init $file $GOOD_KEY
  grep us-east-1 $file
}
