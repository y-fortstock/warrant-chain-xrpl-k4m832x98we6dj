package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"
	tokenv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/token/v1"
	typesv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	LoanAmount       = 1000000 * 100
	LoanCurrency     = "RLUSD"
	LoanInterestRate = 3650
	LoanPeriod       = 10 * time.Minute
)

type Loan struct {
	Amount          uint64 // in Currency * 100
	InterestRate    uint64 // in % * 100
	Period          time.Duration
	NextPaymentDate time.Time
	OwnerWallet     *wallet.Wallet
	CreditorWallet  *wallet.Wallet
	Currency        string
	// LoanEndDate         time.Time
}

func NewLoan(ownerWallet *wallet.Wallet, creditorWallet *wallet.Wallet) Loan {
	return Loan{
		Amount:          LoanAmount,
		Currency:        LoanCurrency,
		InterestRate:    LoanInterestRate,
		Period:          LoanPeriod,
		NextPaymentDate: time.Now().Add(LoanPeriod),
		OwnerWallet:     ownerWallet,
		CreditorWallet:  creditorWallet,
	}
}

type Loans struct {
	loans map[string]Loan
}

func NewLoans() *Loans {
	l := &Loans{loans: make(map[string]Loan)}
	go l.processLoans()

	return l
}

func (l *Loans) AddLoan(tokenID string, loan Loan) {
	l.loans[tokenID] = loan
}

func (l *Loans) RemoveLoan(tokenID string) {
	delete(l.loans, tokenID)
}

func (l *Loans) processLoans() {
	for {
		for tokenID, loan := range l.loans {
			if loan.LoanNextPaymentDate.Before(time.Now()) {
				loan.LoanNextPaymentDate = loan.LoanNextPaymentDate.Add(loan.LoanPeriod)
				l.processLoan(tokenID, loan)
			}
		}
		time.Sleep(10 * time.Minute)
	}
}

func (l *Loans) processLoan(tokenID string, loan Loan) {
	// TODO: implement Interest Tracking here
	// TODO: implement Interest Tracking here
	// TODO: implement Interest Tracking here

	// Owner/Borrower sends payment of RLUSD to Creditor/Lender with interest term
}

func (t *Token) transferToCreditor(ctx context.Context, req *tokenv1.TransferToCreditorRequest) (*tokenv1.TransferToCreditorResponse, error) {
	l := t.logger.With("method", "TransferToCreditor",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
		"owner_address_id", req.GetOwnerAddressId(),
		"token_id", req.GetTokenId(),
	)
	l.Debug("start")
	t.bc.Lock()
	defer t.bc.Unlock()

	creditorSeeds := strings.Split(req.GetCreditorPass(), "-")
	creditor, err := crypto.NewWalletFromHexSeed(creditorSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", creditorSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create recipient wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create recipient wallet: %v", err)
	}
	if !strings.EqualFold(creditor.ClassicAddress.String(), req.GetCreditorAddressId()) {
		l.Error("creditor address does not match", "creditor_address", creditor.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "creditor address does not match")
	}

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

	l.Debug("authorizing token")
	hash, err := t.bc.AuthorizeMPToken(creditor, req.GetTokenId())
	if err != nil {
		l.Warn("failed to authorize token", "error", err)
	}
	l.Debug("authorized token", "hash", hash)

	l.Debug("transferring token to creditor")
	hash, err = t.bc.TransferMPToken(owner, req.GetTokenId(), creditor.ClassicAddress.String())
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

func (t *Token) transferToCreditorWithLoan(ctx context.Context, req *tokenv1.TransferToCreditorRequest) (*tokenv1.TransferToCreditorResponse, error) {
	l := t.logger.With("method", "TransferToCreditorWithLoan",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
		"owner_address_id", req.GetOwnerAddressId(),
		"token_id", req.GetTokenId(),
	)
	l.Debug("start")
	t.bc.Lock()
	defer t.bc.Unlock()

	creditorSeeds := strings.Split(req.GetCreditorPass(), "-")
	creditor, err := crypto.NewWalletFromHexSeed(creditorSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", creditorSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create recipient wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create recipient wallet: %v", err)
	}
	if !strings.EqualFold(creditor.ClassicAddress.String(), req.GetCreditorAddressId()) {
		l.Error("creditor address does not match", "creditor_address", creditor.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "creditor address does not match")
	}

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

	// TODO: implement Deployment of loan here
	// TODO: implement Deployment of loan here
	// TODO: implement Deployment of loan here

	// REPLENISH Owner/Borrower Account with RLUSD (sum of loan interest) from System Account
	// REPLENISH Creditor/Lender Account with RLUSD (loan body) from System Account
	// Owner/Borrower mints Debt Token (with constants in loan terms)
	// Creditor/Lender authorize Debt Token
	// Transfer Debt Token from Owner/Borrower to Creditor/Lender (Creditor/Lender will have 2 tokens Debt and Warrant)
	// Creditor/Lender sends payment of RLUSD to Owner/Borrower with loan term
	// Starting cyclic process of getting interests

	l.Debug("authorizing token")
	hash, err := t.bc.AuthorizeMPToken(creditor, req.GetTokenId())
	if err != nil {
		l.Warn("failed to authorize token", "error", err)
	}
	l.Debug("authorized token", "hash", hash)

	l.Debug("transferring token to creditor")
	hash, err = t.bc.TransferMPToken(owner, req.GetTokenId(), creditor.ClassicAddress.String())
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

func (t *Token) buyoutFromCreditor(ctx context.Context, req *tokenv1.BuyoutFromCreditorRequest) (*tokenv1.BuyoutFromCreditorResponse, error) {
	l := t.logger.With("method", "BuyoutFromCreditor",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
		"owner_address_id", req.GetOwnerAddressId(),
		"token_id", req.GetTokenId(),
	)
	l.Debug("start")
	t.bc.Lock()
	defer t.bc.Unlock()

	creditorSeeds := strings.Split(req.GetCreditorAddressPass(), "-")
	creditor, err := crypto.NewWalletFromHexSeed(creditorSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", creditorSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create recipient wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create recipient wallet: %v", err)
	}
	if !strings.EqualFold(creditor.ClassicAddress.String(), req.GetCreditorAddressId()) {
		l.Error("creditor address does not match", "creditor_address", creditor.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "creditor address does not match")
	}

	ownerSeeds := strings.Split(req.GetOwnerPass(), "-")
	owner, err := crypto.NewWalletFromHexSeed(ownerSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", ownerSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create sender wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create sender wallet: %v", err)
	}
	if !strings.EqualFold(owner.ClassicAddress.String(), req.GetOwnerAddressId()) {
		l.Error("owner address does not match", "owner_address", owner.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "owner address does not match")
	}

	_, err = t.bc.AuthorizeMPToken(owner, req.GetTokenId())
	if err != nil {
		l.Warn("failed to authorize token", "error", err)
	}

	hash, err := t.bc.TransferMPToken(creditor, req.GetTokenId(), owner.ClassicAddress.String())
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

func (t *Token) buyoutFromCreditorWithLoan(ctx context.Context, req *tokenv1.BuyoutFromCreditorRequest) (*tokenv1.BuyoutFromCreditorResponse, error) {
	l := t.logger.With("method", "BuyoutFromCreditorWithLoan",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
		"owner_address_id", req.GetOwnerAddressId(),
		"token_id", req.GetTokenId(),
	)
	l.Debug("start")
	t.bc.Lock()
	defer t.bc.Unlock()

	creditorSeeds := strings.Split(req.GetCreditorAddressPass(), "-")
	creditor, err := crypto.NewWalletFromHexSeed(creditorSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", creditorSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create recipient wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create recipient wallet: %v", err)
	}
	if !strings.EqualFold(creditor.ClassicAddress.String(), req.GetCreditorAddressId()) {
		l.Error("creditor address does not match", "creditor_address", creditor.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "creditor address does not match")
	}

	ownerSeeds := strings.Split(req.GetOwnerPass(), "-")
	owner, err := crypto.NewWalletFromHexSeed(ownerSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", ownerSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create sender wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create sender wallet: %v", err)
	}
	if !strings.EqualFold(owner.ClassicAddress.String(), req.GetOwnerAddressId()) {
		l.Error("owner address does not match", "owner_address", owner.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "owner address does not match")
	}

	// TODO: implement Loan Repayment here
	// TODO: implement Loan Repayment here
	// TODO: implement Loan Repayment here

	// Owner/Borrower sends payment of RLUSD with body of loan (or rest body of loan - there is no implementation reducing body of loan for now) to Creditor/Lender
	// Transfer Debt Token back from Creditor/Lender to Owner/Borrower
	// Stops instance of Interest Tracking for this Debt Token (by apt_issuance_id)
	// Burn Debt Token by Owner/Borrower

	_, err = t.bc.AuthorizeMPToken(owner, req.GetTokenId())
	if err != nil {
		l.Warn("failed to authorize token", "error", err)
	}

	hash, err := t.bc.TransferMPToken(creditor, req.GetTokenId(), owner.ClassicAddress.String())
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

func (t *Token) transferFromCreditorToWarehouse(ctx context.Context, req *tokenv1.TransferFromCreditorToWarehouseRequest) (*tokenv1.TransferFromCreditorToWarehouseResponse, error) {
	l := t.logger.With("method", "TransferFromOwnerToWarehouse",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
		"token_id", req.GetTokenId(),
	)
	l.Debug("start")
	t.bc.Lock()
	defer t.bc.Unlock()

	creditorSeeds := strings.Split(req.GetCreditorAddressPass(), "-")
	creditor, err := crypto.NewWalletFromHexSeed(creditorSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", creditorSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create sender wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create sender wallet: %v", err)
	}
	if !strings.EqualFold(creditor.ClassicAddress.String(), req.GetCreditorAddressId()) {
		l.Error("creditor address does not match", "creditor_address", creditor.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "creditor address does not match")
	}

	issuerAddr, err := t.bc.GetIssuerAddressFromIssuanceID(req.GetTokenId())
	if err != nil {
		l.Error("failed to get issuer address", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get issuer address: %v", err)
	}

	hash, err := t.bc.TransferMPToken(creditor, req.GetTokenId(), issuerAddr)
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

func (t *Token) transferFromCreditorToWarehouseWithLoan(ctx context.Context, req *tokenv1.TransferFromCreditorToWarehouseRequest) (*tokenv1.TransferFromCreditorToWarehouseResponse, error) {
	l := t.logger.With("method", "TransferFromOwnerToWarehouseWithLoan",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
		"token_id", req.GetTokenId(),
	)
	l.Debug("start")
	t.bc.Lock()
	defer t.bc.Unlock()

	creditorSeeds := strings.Split(req.GetCreditorAddressPass(), "-")
	creditor, err := crypto.NewWalletFromHexSeed(creditorSeeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", creditorSeeds[1]))
	if err != nil {
		t.logger.Error("failed to create sender wallet", "error", err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to create sender wallet: %v", err)
	}
	if !strings.EqualFold(creditor.ClassicAddress.String(), req.GetCreditorAddressId()) {
		l.Error("creditor address does not match", "creditor_address", creditor.ClassicAddress.String())
		return nil, status.Errorf(codes.InvalidArgument, "creditor address does not match")
	}

	issuerAddr, err := t.bc.GetIssuerAddressFromIssuanceID(req.GetTokenId())
	if err != nil {
		l.Error("failed to get issuer address", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get issuer address: %v", err)
	}

	// TODO: implement Default here
	// TODO: implement Default here
	// TODO: implement Default here

	// Transfer Debt Token back from Creditor/Lender to Owner/Borrower
	// Stops instance of Interest Tracking for this Debt Token (by apt_issuance_id)
	// Burn Debt Token by Owner/Borrower

	hash, err := t.bc.TransferMPToken(creditor, req.GetTokenId(), issuerAddr)
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
