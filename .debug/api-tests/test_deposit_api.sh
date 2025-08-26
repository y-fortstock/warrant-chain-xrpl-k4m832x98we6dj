#!/bin/bash

# 6b8ae46d5fb069e04b9261469943637499412a51134f1da7c94385d3fedbc66b0f81cbb19dea4907bf3dd7edae5265059b8dbaa6f40795cd0772c487e1bf02ce-0
# rGJdCLYn4ymLBKkLr6rxLbiEzEbjzhQStf

# 6b8ae46d5fb069e04b9261469943637499412a51134f1da7c94385d3fedbc66b0f81cbb19dea4907bf3dd7edae5265059b8dbaa6f40795cd0772c487e1bf02ce-1
# r2LkPbia182R8MKpnBY1kw37PR7SPQC25

IMPORTS=(
  -import-path proto \
  -proto blockchain/account/v1/account_api.proto \
  -proto blockchain/account/v1/account.proto \
  -proto blockchain/types/v1/error.proto \
  -proto blockchain/types/v1/transaction.proto
)

grpcurl -plaintext "${IMPORTS[@]}" \
  -d '{"password": "d83e08bea4d85992c2dd6efb93f070f94f77611d956bbb0594bc0ef29f864ac5cdefdc550f95fd5f84fcb104ad1084532c45b5cd85db071d70395d12a5996bfb-1"}' \
  localhost:8099 blockchain.account.v1.AccountAPI/Create

grpcurl -plaintext "${IMPORTS[@]}" \
  -d '{
    "account_id": "rfnXJ66sZ3HF8Efu82RXawhCbnVE5scDmX",
    "wei_amount": "2000000"
  }' \
  localhost:8099 blockchain.account.v1.AccountAPI/Deposit

sleep 5

grpcurl -plaintext "${IMPORTS[@]}" \
  -d '{
    "account_id": "rfnXJ66sZ3HF8Efu82RXawhCbnVE5scDmX",
    "account_password": "d83e08bea4d85992c2dd6efb93f070f94f77611d956bbb0594bc0ef29f864ac5cdefdc550f95fd5f84fcb104ad1084532c45b5cd85db071d70395d12a5996bfb-1"
  }' \
  localhost:8099 \
  blockchain.account.v1.AccountAPI/ClearBalance
