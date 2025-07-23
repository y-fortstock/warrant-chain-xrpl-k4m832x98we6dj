#!/bin/bash

# Modular gRPC API test script for TokenAPI
# Requires: grpcurl, jq
# Run from project root. gRPC server must be running at localhost:8099

PROTO_DIR="proto/blockchain"
TOKEN_PROTO="${PROTO_DIR}/token/v1/token_api.proto"
TOKEN_PROTO_IMPORT="${PROTO_DIR}/token/v1/token.proto"
TYPES_ERROR_PROTO="${PROTO_DIR}/types/v1/error.proto"
TYPES_TRANSACTION_PROTO="${PROTO_DIR}/types/v1/transaction.proto"

IMPORTS=(
  -import-path proto \
  -proto blockchain/token/v1/token_api.proto \
  -proto blockchain/token/v1/token.proto \
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
        pass=1
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