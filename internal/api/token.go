package api

import (
	"context"

	"log/slog"

	tokenv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/token/v1"
	typesv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Token is an implementation of tokenv1.TokenAPIServer.
type Token struct {
	tokenv1.UnimplementedTokenAPIServer
	logger *slog.Logger
}

// NewToken returns a new Token implementation.
func NewToken(logger *slog.Logger) *Token {
	return &Token{logger: logger}
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
	return nil, status.Errorf(codes.Unimplemented, "method Emission not implemented")
}

// Transfer transfers a token.
func (t *Token) Transfer(ctx context.Context, req *tokenv1.TransferRequest) (*tokenv1.TransferResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Transfer not implemented")
}

// TransferToCreditor transfers a warrant from owner to creditor.
func (t *Token) TransferToCreditor(ctx context.Context, req *tokenv1.TransferToCreditorRequest) (*tokenv1.TransferToCreditorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TransferToCreditor not implemented")
}

// BuyoutFromCreditor transfers a warrant from creditor to owner.
func (t *Token) BuyoutFromCreditor(ctx context.Context, req *tokenv1.BuyoutFromCreditorRequest) (*tokenv1.BuyoutFromCreditorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BuyoutFromCreditor not implemented")
}

// TransferFromOwnerToWarehouse redeems a token, transferring it from owner to warehouse.
func (t *Token) TransferFromOwnerToWarehouse(ctx context.Context, req *tokenv1.TransferFromOwnerToWarehouseRequest) (*tokenv1.TransferFromOwnerToWarehouseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TransferFromOwnerToWarehouse not implemented")
}

// TransferFromCreditorToWarehouse redeems a token, transferring it from creditor to warehouse.
func (t *Token) TransferFromCreditorToWarehouse(ctx context.Context, req *tokenv1.TransferFromCreditorToWarehouseRequest) (*tokenv1.TransferFromCreditorToWarehouseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TransferFromCreditorToWarehouse not implemented")
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
	return nil, status.Errorf(codes.Unimplemented, "method TransactionInfo not implemented")
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
