#!/bin/bash

# Dedicated test script for Deposit API
# Requires: grpcurl, jq
# Run from project root. gRPC server must be running at localhost:8099

IMPORTS=(
  -import-path proto \
  -proto blockchain/account/v1/account_api.proto \
  -proto blockchain/account/v1/account.proto \
  -proto blockchain/types/v1/error.proto \
  -proto blockchain/types/v1/transaction.proto
)

GRPC_SERVER="localhost:8099"

# Test results
TESTS=()
RESULTS=()

run_deposit_test() {
  local test_name="$1"
  local request_json="$2"
  local expected_success="$3"
  local expected_error_pattern="$4"

  echo "Running Deposit test: $test_name"
  local response
  response=$(grpcurl -plaintext -d "$request_json" "${IMPORTS[@]}" "$GRPC_SERVER" "blockchain.account.v1.AccountAPI/Deposit" 2>&1)
  local grpc_status=$?

  local pass=0
  if [ "$expected_success" = "true" ]; then
    # Expect successful response with transaction structure
    if [ $grpc_status -eq 0 ]; then
      if echo "$response" | jq -e '.transaction.id' >/dev/null 2>&1 && \
         echo "$response" | jq -e '.transaction.blockNumber' >/dev/null 2>&1 && \
         echo "$response" | jq -e '.transaction.blockTime' >/dev/null 2>&1; then
        pass=1
        echo "  ✓ Valid transaction response structure"
      else
        echo "  ✗ Invalid transaction response structure"
      fi
    else
      echo "  ✗ gRPC call failed"
    fi
  else
    # Expect error response
    if [ $grpc_status -ne 0 ]; then
      if [ -n "$expected_error_pattern" ]; then
        if echo "$response" | grep -q "$expected_error_pattern"; then
          pass=1
          echo "  ✓ Expected error pattern found: $expected_error_pattern"
        else
          echo "  ✗ Expected error pattern not found: $expected_error_pattern"
        fi
      else
        pass=1
        echo "  ✓ Error response received as expected"
      fi
    else
      echo "  ✗ Expected error but got success response"
    fi
  fi

  if [ $pass -eq 1 ]; then
    echo "[PASS] $test_name"
    RESULTS+=("PASS")
  else
    echo "[FAIL] $test_name"
    echo "  Request: $request_json"
    echo "  Response: $response"
    RESULTS+=("FAIL")
  fi
  TESTS+=("$test_name")
  echo
}

# ---- DEPOSIT TEST CASES ----

# Test 1: Successful deposit with small amount
DEPOSIT_SMALL_REQ='{ "accountId": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX", "weiAmount": "1000000" }'
run_deposit_test "DepositSmallAmount" "$DEPOSIT_SMALL_REQ" "true" ""

# # Test 2: Successful deposit with medium amount
# DEPOSIT_MEDIUM_REQ='{ "accountId": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX", "weiAmount": "5000000" }'
# run_deposit_test "DepositMediumAmount" "$DEPOSIT_MEDIUM_REQ" "true" ""

# Test 3: Deposit with very large amount (should fail if insufficient balance)
DEPOSIT_LARGE_REQ='{ "accountId": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX", "weiAmount": "999999999999999999" }'
run_deposit_test "DepositLargeAmount" "$DEPOSIT_LARGE_REQ" "false" "system account balance is less than drops to transfer"

# Test 4: Deposit with invalid amount format
DEPOSIT_INVALID_REQ='{ "accountId": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX", "weiAmount": "invalid_amount" }'
run_deposit_test "DepositInvalidAmount" "$DEPOSIT_INVALID_REQ" "false" "invalid wei amount"

# Test 5: Deposit with zero amount
DEPOSIT_ZERO_REQ='{ "accountId": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX", "weiAmount": "0" }'
run_deposit_test "DepositZeroAmount" "$DEPOSIT_ZERO_REQ" "true" ""

# Test 6: Deposit with negative amount (as string, should fail parsing)
DEPOSIT_NEGATIVE_REQ='{ "accountId": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX", "weiAmount": "-1000000" }'
run_deposit_test "DepositNegativeAmount" "$DEPOSIT_NEGATIVE_REQ" "false" "invalid wei amount"

# Test 7: Deposit to non-existent account
DEPOSIT_NONEXISTENT_REQ='{ "accountId": "rNonExistentAccount123456789", "weiAmount": "1000000" }'
run_deposit_test "DepositToNonExistentAccount" "$DEPOSIT_NONEXISTENT_REQ" "false" ""

# Test 8: Deposit with missing account ID
DEPOSIT_NO_ACCOUNT_REQ='{ "weiAmount": "1000000" }'
run_deposit_test "DepositNoAccountId" "$DEPOSIT_NO_ACCOUNT_REQ" "false" ""

# Test 9: Deposit with missing amount
DEPOSIT_NO_AMOUNT_REQ='{ "accountId": "ra5nK24KXen9AHvsdFTKHSANinZseWnPcX" }'
run_deposit_test "DepositNoAmount" "$DEPOSIT_NO_AMOUNT_REQ" "false" ""

# ---- SUMMARY ----
echo "==== DEPOSIT API TEST SUMMARY ===="
pass_count=0
fail_count=0
for i in "${!TESTS[@]}"; do
  if [ "${RESULTS[$i]}" = "PASS" ]; then
    ((pass_count++))
  else
    ((fail_count++))
  fi
  echo "${TESTS[$i]}: ${RESULTS[$i]}"
done
echo "----------------------------------"
echo "Total: $((pass_count+fail_count)), Passed: $pass_count, Failed: $fail_count"
if [ $fail_count -eq 0 ]; then
  echo "ALL DEPOSIT TESTS PASSED"
else
  echo "SOME DEPOSIT TESTS FAILED"
fi 