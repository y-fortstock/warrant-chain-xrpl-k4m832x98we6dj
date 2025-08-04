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

# grpcurl -plaintext "${IMPORTS[@]}" \
#   -d '{"password": "6b8ae46d5fb069e04b9261469943637499412a51134f1da7c94385d3fedbc66b0f81cbb19dea4907bf3dd7edae5265059b8dbaa6f40795cd0772c487e1bf02ce-1"}' \
#   localhost:8099 blockchain.account.v1.AccountAPI/Create

grpcurl -plaintext "${IMPORTS[@]}" \
  -d '{
    "account_id": "r2LkPbia182R8MKpnBY1kw37PR7SPQC25",
    "wei_amount": "5000"
  }' \
  localhost:8099 blockchain.account.v1.AccountAPI/Deposit

sleep 5

grpcurl -plaintext "${IMPORTS[@]}" \
  -d '{
    "account_id": "r2LkPbia182R8MKpnBY1kw37PR7SPQC25",
    "account_password": "6b8ae46d5fb069e04b9261469943637499412a51134f1da7c94385d3fedbc66b0f81cbb19dea4907bf3dd7edae5265059b8dbaa6f40795cd0772c487e1bf02ce-1"
  }' \
  localhost:8099 \
  blockchain.account.v1.AccountAPI/ClearBalance
