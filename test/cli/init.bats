#!/usr/bin/env bats
load "conftest"

@test "init fails when no destination is given" {
  [[ "$(bc init)" == *"Destination to save config file missing"* ]]
}
