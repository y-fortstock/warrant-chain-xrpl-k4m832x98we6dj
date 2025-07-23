#!/bin/bash

# This script tests the CreateContract method of the TokenAPI gRPC service using grpcurl.
# The gRPC server is expected to be running and accessible at localhost:8099.
# Make sure grpcurl is installed and available in your PATH.
# The script assumes it is run from the project root.

PROTO_DIR="proto/blockchain"
TOKEN_PROTO="${PROTO_DIR}/token/v1/token_api.proto"
TOKEN_PROTO_IMPORT="${PROTO_DIR}/token/v1/token.proto"
TYPES_ERROR_PROTO="${PROTO_DIR}/types/v1/error.proto"
TYPES_TRANSACTION_PROTO="${PROTO_DIR}/types/v1/transaction.proto"

# Sample request data (edit as needed)
REQUEST_JSON='{ "name": "test_contract" }'

# Run grpcurl to call CreateContract
grpcurl \
  -plaintext \
  -d "$REQUEST_JSON" \
  -import-path proto \
  -proto blockchain/token/v1/token_api.proto \
  -proto blockchain/token/v1/token.proto \
  -proto blockchain/types/v1/error.proto \
  -proto blockchain/types/v1/transaction.proto \
  localhost:8099 \
  blockchain.token.v1.TokenAPI/CreateContract 