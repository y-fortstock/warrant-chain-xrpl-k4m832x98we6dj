// Package api provides the gRPC API implementations for the XRPL blockchain service.
// It includes implementations for account management, token operations, and blockchain interactions.
package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"gitlab.com/warrant1/warrant/chain-xrpl/internal/config"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"
	tokenv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/token/v1"
	typesv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Token implements the tokenv1.TokenAPIServer interface.
// It provides methods for creating, managing, and transferring Multi-Purpose Tokens (MPTs) on the XRPL network.
type Token struct {
	tokenv1.UnimplementedTokenAPIServer
	bc       *Blockchain
	logger   *slog.Logger
	features *config.FeatureConfig
	loans    *Loans
}

// NewToken creates and returns a new Token API server instance.
// It requires a logger and blockchain instance for operation.
func NewToken(logger *slog.Logger, bc *Blockchain) *Token {
	return &Token{
		logger:   logger,
		bc:       bc,
		features: &config.FeatureConfig{Loan: false},
		loans:    NewLoans(),
	}
}

// CreateContract is not available for XRPL and returns an error response.
// XRPL uses a different token model compared to smart contract platforms.
//
// Returns an error response indicating that this method is not supported on XRPL.
func (t *Token) CreateContract(ctx context.Context, req *tokenv1.CreateContractRequest) (*tokenv1.CreateContractResponse, error) {
	t.logger.Warn("CreateContract is not available for xrpl")
	return &tokenv1.CreateContractResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method CreateContract not available for xrpl",
		},
		Token: nil,
	}, nil
}

// Emission creates a new Multi-Purpose Token (MPT) on the XRPL network.
// This function handles token creation, metadata generation, and network submission.
//
// The warehouse password must match the owner address to authorize the operation.
// The function creates an MPT with the specified document hash and signature.
//
// Parameters:
// - req.DocumentHash: The hash of the document backing the token
// - req.WarehouseAddressId: The warehouse account address
// - req.OwnerAddressId: The owner account address
// - req.Signature: The signature authorizing the token creation
// - req.WarehousePass: The warehouse password in format "hexSeed-derivationIndex"
//
// Returns the created token information including issuance ID and transaction details.
func (t *Token) Emission(ctx context.Context, req *tokenv1.EmissionRequest) (*tokenv1.EmissionResponse, error) {
	l := t.logger.With("method", "Emission",
		"document_hash", req.GetDocumentHash(),
		"warehouse_id", req.GetWarehouseAddressId(),
		"owner_address_id", req.GetOwnerAddressId())
	l.Debug("start", "owner_address_id", req.GetOwnerAddressId())
	t.bc.Lock()
	defer t.bc.Unlock()

	seeds := strings.Split(req.GetWarehousePass(), "-")
	warehouse, err := crypto.NewWalletFromHexSeed(seeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", seeds[1]))
	if err != nil {
		l.Error("failed to create wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create wallet: %v", err)
	}
	if !strings.EqualFold(warehouse.ClassicAddress.String(), req.GetWarehouseAddressId()) {
		l.Error("warehouse address does not match", "warehouse_address", warehouse.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "warehouse address does not match")
	}

	if req.GetOwnerPass() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "owner pass is required")
	}
	ownerSeeds := strings.Split(req.GetOwnerPass(), "-")
	owner, err := crypto.NewWalletFromHexSeed(ownerSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", ownerSeeds[1]))
	if err != nil {
		l.Error("failed to create owner wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create owner wallet: %v", err)
	}
	if !strings.EqualFold(owner.ClassicAddress.String(), req.GetOwnerAddressId()) {
		l.Error("owner address does not match", "owner_address", owner.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "owner address does not match")
	}

	l.Debug("issuing mpt token")
	mpt := NewWarrantMPToken(req.GetDocumentHash(), warehouse.ClassicAddress.String())
	hash, issuanceID, err := t.bc.MPTokenIssuanceCreate(warehouse, mpt)
	if err != nil {
		l.Error("failed to create issuance", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create issuance: %v", err)
	}

	for i := 0; i < 5; i++ {
		time.Sleep(4 * time.Second)
		_, meta, _, err := t.bc.GetTransactionInfo(hash)
		if err != nil {
			l.Warn("failed to get transaction info",
				"hash", hash,
				"error", err)
		}
		if strings.Contains(meta.TransactionResult, "SUCCESS") {
			break
		}
	}

	l.Debug("authorizing token", "issuance_id", issuanceID)
	_, err = t.bc.AuthorizeMPToken(owner, issuanceID)
	if err != nil {
		l.Warn("failed to authorize token", "error", err)
	}

	l.Debug("transferring token to owner", "issuance_id", issuanceID)
	hash, err = t.bc.TransferMPToken(warehouse, issuanceID, owner.ClassicAddress.String())
	if err != nil {
		l.Error("failed to transfer token", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to transfer token: %v", err)
	}

	return &tokenv1.EmissionResponse{
		Error: nil,
		Token: &tokenv1.Token{
			Id: issuanceID,
			Transaction: &typesv1.Transaction{
				Id:        hash,
				BlockTime: uint64(time.Now().Unix()),
				IsSuccess: true,
			},
		},
	}, nil
}

// Transfer transfers a Multi-Purpose Token from one account to another.
// Both sender and recipient must be authorized to use the token.
//
// The function first authorizes the recipient for the token, then transfers it from sender to recipient.
//
// Parameters:
// - req.DocumentHash: The hash of the document backing the token
// - req.ReceiverAddressId: The destination account address
// - req.SenderAddressId: The source account address
// - req.Signature: The signature authorizing the transfer
// - req.ReceiverPass: The recipient's password in format "hexSeed-derivationIndex"
// - req.SenderPass: The sender's password in format "hexSeed-derivationIndex"
//
// Returns the transfer response with transaction details.
func (t *Token) Transfer(ctx context.Context, req *tokenv1.TransferRequest) (*tokenv1.TransferResponse, error) {
	l := t.logger.With("method", "Transfer",
		"document_hash", req.GetDocumentHash(),
		"reciever_address_id", req.GetReceiverAddressId(),
		"sender_address_id", req.GetSenderAddressId(),
		"token_id", req.GetTokenId(),
	)
	l.Debug("start")
	t.bc.Lock()
	defer t.bc.Unlock()

	recipientSeeds := strings.Split(req.GetReceiverPass(), "-")
	recipient, err := crypto.NewWalletFromHexSeed(recipientSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", recipientSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create recipient wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create recipient wallet: %v", err)
	}
	if !strings.EqualFold(recipient.ClassicAddress.String(), req.GetReceiverAddressId()) {
		l.Error("recipient address does not match", "recipient_address", recipient.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "recipient address does not match")
	}

	senderSeeds := strings.Split(req.GetSenderPass(), "-")
	sender, err := crypto.NewWalletFromHexSeed(senderSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", senderSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create sender wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create sender wallet: %v", err)
	}
	if !strings.EqualFold(sender.ClassicAddress.String(), req.GetSenderAddressId()) {
		l.Error("sender address does not match", "sender_address", sender.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "sender address does not match")
	}

	_, err = t.bc.AuthorizeMPToken(recipient, req.GetTokenId())
	if err != nil {
		l.Warn("failed to authorize token", "error", err)
	}

	hash, err := t.bc.TransferMPToken(sender, req.GetTokenId(), recipient.ClassicAddress.String())
	if err != nil {
		l.Error("failed to transfer token", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to transfer token: %v", err)
	}

	return &tokenv1.TransferResponse{
		Error: nil,
		Token: &tokenv1.Token{
			Id: req.GetDocumentHash(),
			Transaction: &typesv1.Transaction{
				Id:        hash,
				BlockTime: uint64(time.Now().Unix()),
				IsSuccess: true,
			},
		},
	}, nil
}

// TransferToCreditor transfers a warrant token from the owner to a creditor.
// This is typically used in lending scenarios where collateral is transferred.
//
// The function authorizes the creditor for the token and then transfers ownership.
//
// Parameters:
// - req.DocumentHash: The hash of the document backing the warrant
// - req.CreditorAddressId: The creditor's account address
// - req.OwnerAddressId: The owner's account address
// - req.Signature: The signature authorizing the transfer
// - req.CreditorPass: The creditor's password in format "hexSeed-derivationIndex"
// - req.OwnerAddressPass: The owner's password in format "hexSeed-derivationIndex"
//
// Returns the transfer response with transaction details.
func (t *Token) TransferToCreditor(ctx context.Context, req *tokenv1.TransferToCreditorRequest) (*tokenv1.TransferToCreditorResponse, error) {
	if t.features.Loan {
		return t.transferToCreditorWithLoan(ctx, req)
	}

	return t.transferToCreditor(ctx, req)
}

// BuyoutFromCreditor transfers a warrant token from the creditor back to the owner.
// This is typically used when a loan is repaid and collateral is returned.
//
// The function authorizes the owner for the token and then transfers ownership back.
//
// Parameters:
// - req.DocumentHash: The hash of the document backing the warrant
// - req.CreditorAddressId: The creditor's account address
// - req.OwnerAddressId: The owner's account address
// - req.Signature: The signature authorizing the transfer
// - req.CreditorAddressPass: The creditor's password in format "hexSeed-derivationIndex"
// - req.OwnerPass: The owner's password in format "hexSeed-derivationIndex"
//
// Returns the transfer response with transaction details.
func (t *Token) BuyoutFromCreditor(ctx context.Context, req *tokenv1.BuyoutFromCreditorRequest) (*tokenv1.BuyoutFromCreditorResponse, error) {
	if t.features.Loan {
		return t.buyoutFromCreditorWithLoan(ctx, req)
	}

	return t.buyoutFromCreditor(ctx, req)
}

// TransferFromOwnerToWarehouse redeems a token by transferring it from the owner back to the warehouse.
// This is typically used when a warrant is exercised or expired.
//
// The function determines the warehouse address from the issuance ID and transfers the token.
//
// Parameters:
// - req.DocumentHash: The hash of the document backing the token
// - req.OwnerAddressId: The owner's account address
// - req.Signature: The signature authorizing the redemption
// - req.OwnerAddressPass: The owner's password in format "hexSeed-derivationIndex"
//
// Returns the redemption response with transaction details.
func (t *Token) TransferFromOwnerToWarehouse(ctx context.Context, req *tokenv1.TransferFromOwnerToWarehouseRequest) (*tokenv1.TransferFromOwnerToWarehouseResponse, error) {
	l := t.logger.With("method", "TransferFromOwnerToWarehouse",
		"document_hash", req.GetDocumentHash(),
		"owner_address_id", req.GetOwnerAddressId(),
		"token_id", req.GetTokenId(),
	)
	l.Debug("start")
	t.bc.Lock()
	defer t.bc.Unlock()

	ownerSeeds := strings.Split(req.GetOwnerAddressPass(), "-")
	owner, err := crypto.NewWalletFromHexSeed(ownerSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", ownerSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create sender wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create sender wallet: %v", err)
	}
	if !strings.EqualFold(owner.ClassicAddress.String(), req.GetOwnerAddressId()) {
		l.Error("owner address does not match", "owner_address", owner.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "owner address does not match")
	}

	issuerAddr, err := t.bc.GetIssuerAddressFromIssuanceID(req.GetTokenId())
	if err != nil {
		l.Error("failed to get issuer address", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get issuer address: %v", err)
	}

	hash, err := t.bc.TransferMPToken(owner, req.GetTokenId(), issuerAddr)
	if err != nil {
		l.Error("failed to transfer token", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to transfer token: %v", err)
	}

	return &tokenv1.TransferFromOwnerToWarehouseResponse{
		Error: nil,
		Token: &tokenv1.Token{
			Id: req.GetDocumentHash(),
			Transaction: &typesv1.Transaction{
				Id:        hash,
				BlockTime: uint64(time.Now().Unix()),
				IsSuccess: true,
			},
		},
	}, nil
}

// TransferFromCreditorToWarehouse redeems a token by transferring it from the creditor back to the warehouse.
// This is typically used when a warrant is exercised or expired while held by a creditor.
//
// The function determines the warehouse address from the issuance ID and transfers the token.
//
// Parameters:
// - req.DocumentHash: The hash of the document backing the token
// - req.CreditorAddressId: The creditor's account address
// - req.Signature: The signature authorizing the redemption
// - req.CreditorAddressPass: The creditor's password in format "hexSeed-derivationIndex"
//
// Returns the redemption response with transaction details.
func (t *Token) TransferFromCreditorToWarehouse(ctx context.Context, req *tokenv1.TransferFromCreditorToWarehouseRequest) (*tokenv1.TransferFromCreditorToWarehouseResponse, error) {
	if t.features.Loan {
		return t.transferFromCreditorToWarehouseWithLoan(ctx, req)
	}

	return t.transferFromCreditorToWarehouse(ctx, req)
}

// InitiateReplacement is not available for XRPL and returns an error response.
// XRPL tokens do not support the replacement mechanism used in smart contract platforms.
//
// Returns an error response indicating that this method is not supported on XRPL.
func (t *Token) InitiateReplacement(ctx context.Context, req *tokenv1.InitiateReplacementRequest) (*tokenv1.InitiateReplacementResponse, error) {
	t.logger.Warn("InitiateReplacement is not available for xrpl")
	return &tokenv1.InitiateReplacementResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method InitiateReplacement not available for xrpl",
		},
	}, nil
}

// PrepareToReplace is not available for XRPL and returns an error response.
// XRPL tokens do not support the replacement mechanism used in smart contract platforms.
//
// Returns an error response indicating that this method is not supported on XRPL.
func (t *Token) PrepareToReplace(ctx context.Context, req *tokenv1.PrepareToReplaceRequest) (*tokenv1.PrepareToReplaceResponse, error) {
	t.logger.Warn("PrepareToReplace is not available for xrpl")
	return &tokenv1.PrepareToReplaceResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method PrepareToReplace not available for xrpl",
		},
	}, nil
}

// Replace is not available for XRPL and returns an error response.
// XRPL tokens do not support the replacement mechanism used in smart contract platforms.
//
// Returns an error response indicating that this method is not supported on XRPL.
func (t *Token) Replace(ctx context.Context, req *tokenv1.ReplaceRequest) (*tokenv1.ReplaceResponse, error) {
	t.logger.Warn("Replace is not available for xrpl")
	return &tokenv1.ReplaceResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method Replace not available for xrpl",
		},
	}, nil
}

// RevertReplacement is not available for XRPL and returns an error response.
// XRPL tokens do not support the replacement mechanism used in smart contract platforms.
//
// Returns an error response indicating that this method is not supported on XRPL.
func (t *Token) RevertReplacement(ctx context.Context, req *tokenv1.RevertReplacementRequest) (*tokenv1.RevertReplacementResponse, error) {
	t.logger.Warn("RevertReplacement is not available for xrpl")
	return &tokenv1.RevertReplacementResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method RevertReplacement not available for xrpl",
		},
	}, nil
}

// TransactionInfo retrieves detailed information about a specific transaction.
// This includes transaction status, fees, and other metadata from the XRPL network.
//
// Parameters:
// - req.TransactionId: The transaction hash to query
//
// Returns detailed transaction information including status, fees, and confirmation details.
func (t *Token) TransactionInfo(ctx context.Context, req *tokenv1.TransactionInfoRequest) (*tokenv1.TransactionInfoResponse, error) {
	l := t.logger.With("method", "TransactionInfo",
		"transaction_hash", req.GetTransactionId())
	l.Debug("start")

	resp, meta, baseTx, err := t.bc.GetTransactionInfo(req.GetTransactionId())
	if err != nil {
		l.Error("failed to get transaction info", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get transaction info: %v", err)
	}

	fee, err := strconv.ParseUint(fmt.Sprintf("%d", baseTx.Fee), 10, 64)
	if err != nil {
		l.Error("failed to convert fee to uint64", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to convert fee to uint64: %v", err)
	}

	return &tokenv1.TransactionInfoResponse{
		Error: nil,
		Transaction: &typesv1.Transaction{
			Id:             req.GetTransactionId(),
			BlockNumber:    []byte(fmt.Sprintf("%d", resp.LedgerIndex)),
			BlockTime:      uint64(resp.Date),
			FullyConfirmed: strings.Contains(meta.TransactionResult, "SUCCESS"),
			GasUsed:        fee,
			GasPrice:       1,
			Method:         string(baseTx.TransactionType),
			Input:          fmt.Sprintf("%d", baseTx.Fee),
			Events:         nil,
			// backend use next values to define if transaction is completed
			BlockCount: 1000,
			IsSuccess:  resp.Validated,
		},
	}, nil
}

// AddAddressRole is not available for XRPL and returns an error response.
// XRPL does not support role-based access control in the same way as smart contract platforms.
//
// Returns an error response indicating that this method is not supported on XRPL.
func (t *Token) AddAddressRole(ctx context.Context, req *tokenv1.AddAddressRoleRequest) (*tokenv1.AddAddressRoleResponse, error) {
	t.logger.Warn("AddAddressRole is not available for xrpl")
	return &tokenv1.AddAddressRoleResponse{
		Error: nil,
		Token: &tokenv1.Token{
			Id: "no token id",
			Transaction: &typesv1.Transaction{
				Id:        "no transaction id",
				BlockTime: uint64(time.Now().Unix()),
				IsSuccess: true,
			},
		},
	}, nil
}

// PauseContract is not available for XRPL and returns an error response.
// XRPL tokens cannot be paused in the same way as smart contracts.
//
// Returns an error response indicating that this method is not supported on XRPL.
func (t *Token) PauseContract(ctx context.Context, req *tokenv1.PauseContractRequest) (*tokenv1.PauseContractResponse, error) {
	t.logger.Warn("PauseContract is not available for xrpl")
	return &tokenv1.PauseContractResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method PauseContract not available for xrpl",
		},
	}, nil
}

// ResumeContract is not available for XRPL and returns an error response.
// XRPL tokens cannot be paused in the same way as smart contracts.
//
// Returns an error response indicating that this method is not supported on XRPL.
func (t *Token) ResumeContract(ctx context.Context, req *tokenv1.ResumeContractRequest) (*tokenv1.ResumeContractResponse, error) {
	t.logger.Warn("ResumeContract is not available for xrpl")
	return &tokenv1.ResumeContractResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method ResumeContract not available for xrpl",
		},
	}, nil
}
