#!/bin/bash

# Configuration variables
ACCOUNT_ID="rfnXJ66sZ3HF8Efu82RXawhCbnVE5scDmX"
PASSWORD="d83e08bea4d85992c2dd6efb93f070f94f77611d956bbb0594bc0ef29f864ac5cdefdc550f95fd5f84fcb104ad1084532c45b5cd85db071d70395d12a5996bfb-1"

IMPORTS=(
  -import-path proto \
  -proto blockchain/account/v1/account_api.proto \
  -proto blockchain/account/v1/account.proto \
  -proto blockchain/types/v1/error.proto \
  -proto blockchain/types/v1/transaction.proto
)

grpcurl -plaintext "${IMPORTS[@]}" \
  -d "{\"password\": \"$PASSWORD\"}" \
  localhost:8099 blockchain.account.v1.AccountAPI/Create

grpcurl -plaintext "${IMPORTS[@]}" \
  -d "{
    \"account_id\": \"$ACCOUNT_ID\",
    \"wei_amount\": \"2000000\"
  }" \
  localhost:8099 blockchain.account.v1.AccountAPI/Deposit

sleep 5

grpcurl -plaintext "${IMPORTS[@]}" \
  -d "{
    \"account_id\": \"$ACCOUNT_ID\",
    \"account_password\": \"$PASSWORD\"
  }" \
  localhost:8099 \
  blockchain.account.v1.AccountAPI/ClearBalance
