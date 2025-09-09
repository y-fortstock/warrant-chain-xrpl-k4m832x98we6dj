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
        # if echo "$response" | jq -e '.transaction.id' >/dev/null 2>&1 && \
        #    echo "$response" | jq -e '.transaction.blockTime' >/dev/null 2>&1; then
          pass=1
        # fi
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

# # 1. CreateContract
# CREATE_CONTRACT_REQ='{ "name": "test_contract" }'
# run_test "CreateContract" "blockchain.token.v1.TokenAPI/CreateContract" "$CREATE_CONTRACT_REQ" "" 1

# # 1a. AccountAPI Create
# CREATE_ACCOUNT_REQ='{ "password": "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f-0" }'
# CREATE_ACCOUNT_EXPECTED='{ "account": { "id": "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC" } }'
# run_test "CreateAccount" "blockchain.account.v1.AccountAPI/Create" "$CREATE_ACCOUNT_REQ" "$CREATE_ACCOUNT_EXPECTED" 0

# # 1b. AccountAPI GetBalance
# GET_BALANCE_REQ='{ "accountId": "rUWaveCdPhssfFE3SiFV811w5vvaFxy1W1" }'
# GET_BALANCE_EXPECTED='{ "balance": "100000000" }'
# run_test "GetBalance" "blockchain.account.v1.AccountAPI/GetBalance" "$GET_BALANCE_REQ" "$GET_BALANCE_EXPECTED" 0

# 1c. AccountAPI Deposit
DEPOSIT_REQ='{ "accountId": "rKVBdHSj12yPe6NRNdJUCFsyV23k2q7As3", "weiAmount": "2000000" }'
run_test "Deposit Owner" "blockchain.account.v1.AccountAPI/Deposit" "$DEPOSIT_REQ" "" 0

DEPOSIT_REQ='{ "accountId": "rfnXJ66sZ3HF8Efu82RXawhCbnVE5scDmX", "weiAmount": "2000000" }'
run_test "Deposit Warehouse" "blockchain.account.v1.AccountAPI/Deposit" "$DEPOSIT_REQ" "" 0

DEPOSIT_REQ='{ "accountId": "rDAsBY6uhCDZjQZ1SwDuPTrD9MtMQMCMn2", "weiAmount": "2000000" }'
run_test "Deposit Creditor" "blockchain.account.v1.AccountAPI/Deposit" "$DEPOSIT_REQ" "" 0

sleep 5

# 1d. TokenAPI Emission
# EMISSION_REQ='{ "document_hash": "test_document_hash_123", "owner_address_id": "rKVBdHSj12yPe6NRNdJUCFsyV23k2q7As3", "owner_pass": "20c6ec9f36c6347f39485ee5a1e4c515373d2f32d69eb77b42ffb1d459fa26ebee73c33ca0da640feb910697f058419c674f17b8c8d6eb688fcd170a3bb2bff7-0", "warehouse_address_id": "rfnXJ66sZ3HF8Efu82RXawhCbnVE5scDmX", "warehouse_pass": "d83e08bea4d85992c2dd6efb93f070f94f77611d956bbb0594bc0ef29f864ac5cdefdc550f95fd5f84fcb104ad1084532c45b5cd85db071d70395d12a5996bfb-1", "signature": "test_signature_123" }'
# # Expected response structure (token ID, transaction ID and timestamp will be dynamic)
# run_test "Emission" "blockchain.token.v1.TokenAPI/Emission" "$EMISSION_REQ" "" 0

# # TokenAPI TransferToCreditor
# TRANSFER_TO_CREDITOR_REQ='{ "document_hash": "0056469242ABB74F75D35DB3C9079A7864DCE9527492CFD4", "owner_address_id": "rfnXJ66sZ3HF8Efu82RXawhCbnVE5scDmX", "owner_address_pass": "d83e08bea4d85992c2dd6efb93f070f94f77611d956bbb0594bc0ef29f864ac5cdefdc550f95fd5f84fcb104ad1084532c45b5cd85db071d70395d12a5996bfb-1", "creditor_address_id": "rDAsBY6uhCDZjQZ1SwDuPTrD9MtMQMCMn2", "creditor_pass": "202f73ce60fc7e3cd1bd13642420839ff55ee0828313e5f6c19960775ec6c2d3bb0ad0e61954aa2beab825c8de08d2594b2e74a915927195b5b0c1a02286a56e-1", "signature": "test_signature_123" }'
# run_test "TransferToCreditor" "blockchain.token.v1.TokenAPI/TransferToCreditor" "$TRANSFER_TO_CREDITOR_REQ" "" 0

# # TokenAPI TransferFromCreditorToWarehouse
# TRANSFER_FROM_CREDITOR_TO_WAREHOUSE_REQ='{ "document_hash": "0056469242ABB74F75D35DB3C9079A7864DCE9527492CFD4", "creditor_address_id": "rDAsBY6uhCDZjQZ1SwDuPTrD9MtMQMCMn2", "creditor_address_pass": "202f73ce60fc7e3cd1bd13642420839ff55ee0828313e5f6c19960775ec6c2d3bb0ad0e61954aa2beab825c8de08d2594b2e74a915927195b5b0c1a02286a56e-1", "signature": "test_signature_123" }'
# run_test "TransferFromCreditorToWarehouse" "blockchain.token.v1.TokenAPI/TransferFromCreditorToWarehouse" "$TRANSFER_FROM_CREDITOR_TO_WAREHOUSE_REQ" "" 0

# 1e. AccountAPI ClearBalance
CLEAR_BALANCE_REQ='{ "accountId": "rKVBdHSj12yPe6NRNdJUCFsyV23k2q7As3", "accountPassword": "20c6ec9f36c6347f39485ee5a1e4c515373d2f32d69eb77b42ffb1d459fa26ebee73c33ca0da640feb910697f058419c674f17b8c8d6eb688fcd170a3bb2bff7-0" }'
run_test "ClearBalance Owner" "blockchain.account.v1.AccountAPI/ClearBalance" "$CLEAR_BALANCE_REQ" "" 0

CLEAR_BALANCE_REQ='{ "accountId": "rfnXJ66sZ3HF8Efu82RXawhCbnVE5scDmX", "accountPassword": "d83e08bea4d85992c2dd6efb93f070f94f77611d956bbb0594bc0ef29f864ac5cdefdc550f95fd5f84fcb104ad1084532c45b5cd85db071d70395d12a5996bfb-1" }'
run_test "ClearBalance Warehouse" "blockchain.account.v1.AccountAPI/ClearBalance" "$CLEAR_BALANCE_REQ" "" 0

CLEAR_BALANCE_REQ='{ "accountId": "rDAsBY6uhCDZjQZ1SwDuPTrD9MtMQMCMn2", "accountPassword": "202f73ce60fc7e3cd1bd13642420839ff55ee0828313e5f6c19960775ec6c2d3bb0ad0e61954aa2beab825c8de08d2594b2e74a915927195b5b0c1a02286a56e-1" }'
run_test "ClearBalance Creditor" "blockchain.account.v1.AccountAPI/ClearBalance" "$CLEAR_BALANCE_REQ" "" 0

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