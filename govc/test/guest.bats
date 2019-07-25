#!/usr/bin/env bats

load test_helper

@test "guest operations" {
  vcsim_env

  export GOVC_VM=DC0_H0_VM0

  run govc guest.start /bin/df -h "> $BATS_TMPDIR/df.txt"
  assert_success
  pid="$output"

  run govc guest.ps -p "$pid"
  assert_success
  assert_matches /bin/df

  run govc guest.ps -x
  assert_success

  run govc guest.kill -p "$pid"
  assert_success
}
