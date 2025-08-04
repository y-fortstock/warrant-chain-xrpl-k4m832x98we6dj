#!/bin/bash

# Modular gRPC API test script for TokenAPI
# Requires: grpcurl, jq
# Run from project root. gRPC server must be running at localhost:8099
IMPORTS=(
  -import-path proto \
  -proto blockchain/token/v1/token_api.proto \
  -proto blockchain/token/v1/token.proto \
  -proto blockchain/account/v1/account_api.proto \
  -proto blockchain/account/v1/account.proto \
  -proto blockchain/types/v1/error.proto \
  -proto blockchain/types/v1/transaction.proto
)

GRPC_SERVER="localhost:8099"

# Test results
TESTS=()
RESULTS=()

run_test() {
  local test_name="$1"
  local method="$2"
  local request_json="$3"
  local expected_json="$4"
  local expect_unimplemented="$5"

  echo "Running test: $test_name"
  local response
  response=$(grpcurl -plaintext -d "$request_json" "${IMPORTS[@]}" "$GRPC_SERVER" "$method" 2>&1)
  local grpc_status=$?

  local pass=0
  if [ "$expect_unimplemented" = "1" ]; then
    # Accept Unimplemented or not available for xrpl as PASS
    if echo "$response" | grep -q -e 'Unimplemented' -e 'not available for xrpl'; then
      pass=1
    fi
  else
    # Compare using jq (normalize JSON)
    if [ $grpc_status -eq 0 ]; then
      if [ -z "$expected_json" ]; then
        # For dynamic responses, validate structure
        if echo "$response" | jq -e '.transaction.id' >/dev/null 2>&1 && \
           echo "$response" | jq -e '.transaction.blockNumber' >/dev/null 2>&1 && \
           echo "$response" | jq -e '.transaction.blockTime' >/dev/null 2>&1; then
          pass=1
        fi
      else
        diff=$(diff <(echo "$response" | jq -S .) <(echo "$expected_json" | jq -S .))
        if [ -z "$diff" ]; then
          pass=1
        fi
      fi
    fi
  fi

  if [ $pass -eq 1 ]; then
    echo "[PASS] $test_name"
    RESULTS+=("PASS")
  else
    echo "[FAIL] $test_name"
    echo "  Request: $request_json"
    echo "  Expected: $expected_json"
    echo "  Actual:   $response"
    RESULTS+=("FAIL")
  fi
  TESTS+=("$test_name")
  echo
}

# ---- TEST CASES ----

# 1. CreateContract
CREATE_CONTRACT_REQ='{ "name": "test_contract" }'
run_test "CreateContract" "blockchain.token.v1.TokenAPI/CreateContract" "$CREATE_CONTRACT_REQ" "" 1

# 1a. AccountAPI Create
CREATE_ACCOUNT_REQ='{ "password": "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f-0" }'
CREATE_ACCOUNT_EXPECTED='{ "account": { "id": "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC" } }'
run_test "CreateAccount" "blockchain.account.v1.AccountAPI/Create" "$CREATE_ACCOUNT_REQ" "$CREATE_ACCOUNT_EXPECTED" 0

# 1b. AccountAPI GetBalance
GET_BALANCE_REQ='{ "accountId": "rUWaveCdPhssfFE3SiFV811w5vvaFxy1W1" }'
GET_BALANCE_EXPECTED='{ "balance": "10000000" }'
run_test "GetBalance" "blockchain.account.v1.AccountAPI/GetBalance" "$GET_BALANCE_REQ" "$GET_BALANCE_EXPECTED" 0

# # 1c. AccountAPI Deposit
# DEPOSIT_REQ='{ "accountId": "rUWaveCdPhssfFE3SiFV811w5vvaFxy1W1", "weiAmount": "1000000" }'
# # Expected response structure (transaction ID and timestamp will be dynamic)
# DEPOSIT_EXPECTED_STRUCTURE='{ "transaction": { "id": ".*", "blockNumber": "AA==", "blockTime": "[0-9]+" } }'
# run_test "Deposit" "blockchain.account.v1.AccountAPI/Deposit" "$DEPOSIT_REQ" "" 0

# # 1d. AccountAPI Deposit - Large Amount (should fail if insufficient balance)
# DEPOSIT_LARGE_REQ='{ "accountId": "rUWaveCdPhssfFE3SiFV811w5vvaFxy1W1", "weiAmount": "999999999999999999" }'
# run_test "DepositLargeAmount" "blockchain.account.v1.AccountAPI/Deposit" "$DEPOSIT_LARGE_REQ" "" 0

# # 1e. AccountAPI Deposit - Invalid Amount (should fail with parsing error)
# DEPOSIT_INVALID_REQ='{ "accountId": "rUWaveCdPhssfFE3SiFV811w5vvaFxy1W1", "weiAmount": "invalid_amount" }'
# run_test "DepositInvalidAmount" "blockchain.account.v1.AccountAPI/Deposit" "$DEPOSIT_INVALID_REQ" "" 0

# 2. PauseContract
PAUSE_CONTRACT_REQ='{}'
run_test "PauseContract" "blockchain.token.v1.TokenAPI/PauseContract" "$PAUSE_CONTRACT_REQ" "" 1

# 3. ResumeContract
RESUME_CONTRACT_REQ='{}'
run_test "ResumeContract" "blockchain.token.v1.TokenAPI/ResumeContract" "$RESUME_CONTRACT_REQ" "" 1

# ---- SUMMARY ----
echo "==== TEST SUMMARY ===="
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
echo "----------------------"
echo "Total: $((pass_count+fail_count)), Passed: $pass_count, Failed: $fail_count"
if [ $fail_count -eq 0 ]; then
  echo "ALL TESTS PASSED"
else
  echo "SOME TESTS FAILED"
fi