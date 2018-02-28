load lib/lib

@test "login as Alice" {
  run aptomi::login alice
  assert_equal 1 $status
}

@test "addition using dc" {
  [ "$BATS_TEST_LAST_STATUS" -eq "0" ]
  result="$(echo 2 2+p | dc)"
  [ "$result" -eq 4 ]
}
