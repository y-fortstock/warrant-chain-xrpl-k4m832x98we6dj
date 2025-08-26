package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"
	tokenv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/token/v1"
	typesv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Token is an implementation of tokenv1.TokenAPIServer.
type Token struct {
	tokenv1.UnimplementedTokenAPIServer
	bc     *Blockchain
	logger *slog.Logger
}

// NewToken returns a new Token implementation.
func NewToken(logger *slog.Logger, bc *Blockchain) *Token {
	return &Token{logger: logger, bc: bc}
}

// CreateContract creates a smart contract.
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

// Emission emits a token.
func (t *Token) Emission(ctx context.Context, req *tokenv1.EmissionRequest) (*tokenv1.EmissionResponse, error) {
	l := t.logger.With("method", "Emission",
		"document_hash", req.GetDocumentHash(),
		"warehouse_id", req.GetWarehouseAddressId())
	l.Debug("start",
		"owner_address_id", req.GetOwnerAddressId(),
		"signature", req.GetSignature())

	seeds := strings.Split(req.GetWarehousePass(), "-")
	w, err := crypto.NewWalletFromHexSeed(seeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", seeds[1]))
	if err != nil {
		l.Error("failed to create wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create wallet: %v", err)
	}
	if strings.ToLower(string(w.Address)) != strings.ToLower(req.GetOwnerAddressId()) {
		l.Error("warehouse address does not match", "warehouse_address", string(w.Address))
		return nil, status.Errorf(codes.InvalidArgument, "warehouse address does not match")
	}

	mpt := NewMPToken(req.GetDocumentHash(), req.GetSignature())
	hash, issuanceID, err := t.bc.MPTokenIssuanceCreate(w, mpt)
	if err != nil {
		l.Error("failed to create issuance", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create issuance: %v", err)
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

// Transfer transfers a token.
func (t *Token) Transfer(ctx context.Context, req *tokenv1.TransferRequest) (*tokenv1.TransferResponse, error) {
	l := t.logger.With("method", "Transfer",
		"document_hash", req.GetDocumentHash(),
		"reciever_address_id", req.GetReceiverAddressId(),
		"sender_address_id", req.GetSenderAddressId(),
	)
	l.Debug("start", "signature", req.GetSignature())

	recipientSeeds := strings.Split(req.GetReceiverPass(), "-")
	recipient, err := crypto.NewWalletFromHexSeed(recipientSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", recipientSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create recipient wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create recipient wallet: %v", err)
	}
	if strings.ToLower(string(recipient.Address)) != strings.ToLower(req.GetReceiverAddressId()) {
		l.Error("recipient address does not match", "recipient_address", string(recipient.Address))
		return nil, status.Errorf(codes.InvalidArgument, "recipient address does not match")
	}

	senderSeeds := strings.Split(req.GetSenderPass(), "-")
	sender, err := crypto.NewWalletFromHexSeed(senderSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", senderSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create sender wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create sender wallet: %v", err)
	}
	if strings.ToLower(string(sender.Address)) != strings.ToLower(req.GetSenderAddressId()) {
		l.Error("sender address does not match", "sender_address", string(sender.Address))
		return nil, status.Errorf(codes.InvalidArgument, "sender address does not match")
	}

	_, err = t.bc.AuthorizeMPToken(recipient, req.GetDocumentHash())
	if err != nil {
		l.Warn("failed to authorize token", "error", err)
	}

	hash, err := t.bc.TransferMPToken(sender, req.GetDocumentHash(), string(recipient.Address))
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

// TransferToCreditor transfers a warrant from owner to creditor.
func (t *Token) TransferToCreditor(ctx context.Context, req *tokenv1.TransferToCreditorRequest) (*tokenv1.TransferToCreditorResponse, error) {
	l := t.logger.With("method", "TransferToCreditor",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
		"owner_address_id", req.GetOwnerAddressId(),
	)
	l.Debug("start", "signature", req.GetSignature())

	creditorSeeds := strings.Split(req.GetCreditorPass(), "-")
	creditor, err := crypto.NewWalletFromHexSeed(creditorSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", creditorSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create recipient wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create recipient wallet: %v", err)
	}
	if strings.ToLower(string(creditor.Address)) != strings.ToLower(req.GetCreditorAddressId()) {
		l.Error("creditor address does not match", "creditor_address", string(creditor.Address))
		return nil, status.Errorf(codes.InvalidArgument, "creditor address does not match")
	}

	ownerSeeds := strings.Split(req.GetOwnerAddressPass(), "-")
	owner, err := crypto.NewWalletFromHexSeed(ownerSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", ownerSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create sender wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create sender wallet: %v", err)
	}
	if strings.ToLower(string(owner.Address)) != strings.ToLower(req.GetOwnerAddressId()) {
		l.Error("owner address does not match", "owner_address", string(owner.Address))
		return nil, status.Errorf(codes.InvalidArgument, "owner address does not match")
	}

	_, err = t.bc.AuthorizeMPToken(creditor, req.GetDocumentHash())
	if err != nil {
		l.Warn("failed to authorize token", "error", err)
	}

	hash, err := t.bc.TransferMPToken(owner, req.GetDocumentHash(), string(creditor.Address))
	if err != nil {
		l.Error("failed to transfer token", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to transfer token: %v", err)
	}

	return &tokenv1.TransferToCreditorResponse{
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

// BuyoutFromCreditor transfers a warrant from creditor to owner.
func (t *Token) BuyoutFromCreditor(ctx context.Context, req *tokenv1.BuyoutFromCreditorRequest) (*tokenv1.BuyoutFromCreditorResponse, error) {
	l := t.logger.With("method", "BuyoutFromCreditor",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
		"owner_address_id", req.GetOwnerAddressId(),
	)
	l.Debug("start", "signature", req.GetSignature())

	creditorSeeds := strings.Split(req.GetCreditorAddressPass(), "-")
	creditor, err := crypto.NewWalletFromHexSeed(creditorSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", creditorSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create recipient wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create recipient wallet: %v", err)
	}
	if strings.ToLower(string(creditor.Address)) != strings.ToLower(req.GetCreditorAddressId()) {
		l.Error("creditor address does not match", "creditor_address", string(creditor.Address))
		return nil, status.Errorf(codes.InvalidArgument, "creditor address does not match")
	}

	ownerSeeds := strings.Split(req.GetOwnerPass(), "-")
	owner, err := crypto.NewWalletFromHexSeed(ownerSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", ownerSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create sender wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create sender wallet: %v", err)
	}
	if strings.ToLower(string(owner.Address)) != strings.ToLower(req.GetOwnerAddressId()) {
		l.Error("owner address does not match", "owner_address", string(owner.Address))
		return nil, status.Errorf(codes.InvalidArgument, "owner address does not match")
	}

	_, err = t.bc.AuthorizeMPToken(owner, req.GetDocumentHash())
	if err != nil {
		l.Warn("failed to authorize token", "error", err)
	}

	hash, err := t.bc.TransferMPToken(creditor, req.GetDocumentHash(), string(owner.Address))
	if err != nil {
		l.Error("failed to transfer token", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to transfer token: %v", err)
	}

	return &tokenv1.BuyoutFromCreditorResponse{
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

// TransferFromOwnerToWarehouse redeems a token, transferring it from owner to warehouse.
func (t *Token) TransferFromOwnerToWarehouse(ctx context.Context, req *tokenv1.TransferFromOwnerToWarehouseRequest) (*tokenv1.TransferFromOwnerToWarehouseResponse, error) {
	l := t.logger.With("method", "TransferFromOwnerToWarehouse",
		"document_hash", req.GetDocumentHash(),
		"owner_address_id", req.GetOwnerAddressId(),
	)
	l.Debug("start", "signature", req.GetSignature())

	ownerSeeds := strings.Split(req.GetOwnerAddressPass(), "-")
	owner, err := crypto.NewWalletFromHexSeed(ownerSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", ownerSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create sender wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create sender wallet: %v", err)
	}
	if strings.ToLower(string(owner.Address)) != strings.ToLower(req.GetOwnerAddressId()) {
		l.Error("owner address does not match", "owner_address", string(owner.Address))
		return nil, status.Errorf(codes.InvalidArgument, "owner address does not match")
	}

	issuerAddr, err := t.bc.GetIssuerAddressFromIssuanceID(req.GetDocumentHash())
	if err != nil {
		l.Error("failed to get issuer address", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get issuer address: %v", err)
	}

	hash, err := t.bc.TransferMPToken(owner, req.GetDocumentHash(), issuerAddr)
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

// TransferFromCreditorToWarehouse redeems a token, transferring it from creditor to warehouse.
func (t *Token) TransferFromCreditorToWarehouse(ctx context.Context, req *tokenv1.TransferFromCreditorToWarehouseRequest) (*tokenv1.TransferFromCreditorToWarehouseResponse, error) {
	l := t.logger.With("method", "TransferFromOwnerToWarehouse",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
	)
	l.Debug("start", "signature", req.GetSignature())

	creditorSeeds := strings.Split(req.GetCreditorAddressPass(), "-")
	creditor, err := crypto.NewWalletFromHexSeed(creditorSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", creditorSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create sender wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create sender wallet: %v", err)
	}
	if strings.ToLower(string(creditor.Address)) != strings.ToLower(req.GetCreditorAddressId()) {
		l.Error("creditor address does not match", "creditor_address", string(creditor.Address))
		return nil, status.Errorf(codes.InvalidArgument, "creditor address does not match")
	}

	issuerAddr, err := t.bc.GetIssuerAddressFromIssuanceID(req.GetDocumentHash())
	if err != nil {
		l.Error("failed to get issuer address", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get issuer address: %v", err)
	}

	hash, err := t.bc.TransferMPToken(creditor, req.GetDocumentHash(), issuerAddr)
	if err != nil {
		l.Error("failed to transfer token", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to transfer token: %v", err)
	}

	return &tokenv1.TransferFromCreditorToWarehouseResponse{
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

// InitiateReplacement initiates a replacement.
func (t *Token) InitiateReplacement(ctx context.Context, req *tokenv1.InitiateReplacementRequest) (*tokenv1.InitiateReplacementResponse, error) {
	t.logger.Warn("InitiateReplacement is not available for xrpl")
	return &tokenv1.InitiateReplacementResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method InitiateReplacement not available for xrpl",
		},
	}, nil
}

// PrepareToReplace prepares to replace.
func (t *Token) PrepareToReplace(ctx context.Context, req *tokenv1.PrepareToReplaceRequest) (*tokenv1.PrepareToReplaceResponse, error) {
	t.logger.Warn("PrepareToReplace is not available for xrpl")
	return &tokenv1.PrepareToReplaceResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method PrepareToReplace not available for xrpl",
		},
	}, nil
}

// Replace replaces a token.
func (t *Token) Replace(ctx context.Context, req *tokenv1.ReplaceRequest) (*tokenv1.ReplaceResponse, error) {
	t.logger.Warn("Replace is not available for xrpl")
	return &tokenv1.ReplaceResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method Replace not available for xrpl",
		},
	}, nil
}

// RevertReplacement reverts a replacement.
func (t *Token) RevertReplacement(ctx context.Context, req *tokenv1.RevertReplacementRequest) (*tokenv1.RevertReplacementResponse, error) {
	t.logger.Warn("RevertReplacement is not available for xrpl")
	return &tokenv1.RevertReplacementResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method RevertReplacement not available for xrpl",
		},
	}, nil
}

// TransactionInfo returns transaction info.
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
			// BlockCount: 1,
			Id:             req.GetTransactionId(),
			BlockNumber:    []byte(fmt.Sprintf("%d", resp.LedgerIndex)),
			BlockTime:      uint64(resp.Date),
			FullyConfirmed: strings.Contains(meta.TransactionResult, "SUCCESS"),
			IsSuccess:      resp.Validated,
			GasUsed:        fee,
			GasPrice:       1,
			Method:         string(baseTx.TransactionType),
			Input:          fmt.Sprintf("%d", baseTx.Fee),
			Events:         nil,
		},
	}, nil
}

// AddAddressRole sets an address role.
func (t *Token) AddAddressRole(ctx context.Context, req *tokenv1.AddAddressRoleRequest) (*tokenv1.AddAddressRoleResponse, error) {
	t.logger.Warn("AddAddressRole is not available for xrpl")
	return &tokenv1.AddAddressRoleResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method AddAddressRole not available for xrpl",
		},
	}, nil
}

// PauseContract pauses the contract.
func (t *Token) PauseContract(ctx context.Context, req *tokenv1.PauseContractRequest) (*tokenv1.PauseContractResponse, error) {
	t.logger.Warn("PauseContract is not available for xrpl")
	return &tokenv1.PauseContractResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method PauseContract not available for xrpl",
		},
	}, nil
}

// ResumeContract resumes the contract.
func (t *Token) ResumeContract(ctx context.Context, req *tokenv1.ResumeContractRequest) (*tokenv1.ResumeContractResponse, error) {
	t.logger.Warn("ResumeContract is not available for xrpl")
	return &tokenv1.ResumeContractResponse{
		Error: &typesv1.Error{
			Code:        typesv1.Err_ERR_INVALID,
			Description: "method ResumeContract not available for xrpl",
		},
	}, nil
}
