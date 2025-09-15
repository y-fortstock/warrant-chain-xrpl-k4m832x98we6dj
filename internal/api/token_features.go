package api

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/shopspring/decimal"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"
	tokenv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/token/v1"
	typesv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/types/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Loan struct {
	Principal          decimal.Decimal
	AnnualInterestRate decimal.Decimal
	Period             time.Duration
	NextPaymentDate    time.Time
	OwnerWallet        *wallet.Wallet
	CreditorWallet     *wallet.Wallet
	Currency           string
	DebtTokenID        string
	// LoanEndDate         time.Time
}

func NewLoan(ownerWallet *wallet.Wallet, creditorWallet *wallet.Wallet) Loan {
	return Loan{
		Principal:          decimal.NewFromInt(LoanAmount),
		AnnualInterestRate: decimal.NewFromFloat(LoanInterestRate),
		Period:             LoanPeriod,
		NextPaymentDate:    time.Now().Add(LoanPeriod),
		OwnerWallet:        ownerWallet,
		CreditorWallet:     creditorWallet,
		Currency:           LoanCurrency,
	}
}

func (l *Loan) SetDebtTokenID(debtTokenID string) {
	l.DebtTokenID = debtTokenID
}

type Loans struct {
	loans  map[string]Loan
	bc     *Blockchain
	logger *slog.Logger
}

func NewLoans(logger *slog.Logger, bc *Blockchain) *Loans {
	l := &Loans{loans: make(map[string]Loan), logger: logger.With("method", "Loans"), bc: bc}
	go l.processLoans()
	l.logger.Debug("loans initialized and started processing")

	return l
}

func (l *Loans) AddLoan(tokenID string, loan Loan) {
	l.loans[tokenID] = loan
}

func (l *Loans) GetLoan(tokenID string) (Loan, error) {
	loan, ok := l.loans[tokenID]
	if !ok {
		return Loan{}, fmt.Errorf("loan not found")
	}
	return loan, nil
}

func (l *Loans) RemoveLoan(tokenID string) {
	delete(l.loans, tokenID)
}

func (l *Loans) processLoans() {
	for {
		l.logger.Debug("processing loans")
		for tokenID, loan := range l.loans {
			if loan.NextPaymentDate.Before(time.Now()) {
				loan.NextPaymentDate = loan.NextPaymentDate.Add(loan.Period)
				l.loans[tokenID] = loan

				l.logger.Debug("processing loan",
					"token_id", tokenID,
					"next_payment_date", loan.NextPaymentDate,
					"principal", loan.Principal,
					"annual_interest_rate", loan.AnnualInterestRate,
					"period", loan.Period,
					"owner_wallet", loan.OwnerWallet.ClassicAddress.String(),
					"creditor_wallet", loan.CreditorWallet.ClassicAddress.String(),
					"currency", loan.Currency,
				)
				err := l.processLoan(tokenID, loan)
				if err != nil {
					l.logger.Error("failed to process loan", "error", err)
				}
			}
		}
		time.Sleep(time.Minute)
	}
}

func (l *Loans) processLoan(tokenID string, loan Loan) error {
	l.bc.Lock()
	defer l.bc.Unlock()

	dailyRate := loan.AnnualInterestRate.Div(decimal.NewFromInt(100)).Div(decimal.NewFromInt(365))
	interest := loan.Principal.Mul(dailyRate)

	err := l.bc.PaymentRLUSD(loan.OwnerWallet, loan.CreditorWallet, interest.InexactFloat64())
	if err != nil {
		return fmt.Errorf("failed to payment RLUSD: %v", err)
	}
	l.logger.Debug("processed loan", "token_id", tokenID)
	return nil
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
	err = t.bc.AuthorizeMPToken(creditor, req.GetTokenId())
	if err != nil {
		l.Warn("failed to authorize token", "error", err)
	}
	l.Debug("authorized token")

	l.Debug("transferring token to creditor")
	hash, err := t.bc.TransferMPToken(owner, req.GetTokenId(), creditor.ClassicAddress.String())
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
	tokenID := req.GetTokenId()
	l := t.logger.With("method", "TransferToCreditorWithLoan",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
		"owner_address_id", req.GetOwnerAddressId(),
		"token_id", tokenID,
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

	l.Debug("setup initial balances for parties")
	err = t.bc.SystemAccountInit()
	if err != nil {
		l.Error("failed to initialize system account", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to initialize system account: %v", err)
	}

	loan := NewLoan(owner, creditor)

	err = t.bc.CreateTrustlineFromSystemAccount(owner, loan.Principal.InexactFloat64()*10)
	if err != nil {
		l.Error("failed to create trustline", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create trustline: %v", err)
	}

	err = t.bc.CreateTrustlineFromSystemAccount(creditor, loan.Principal.InexactFloat64()*10)
	if err != nil {
		l.Error("failed to create trustline", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create trustline: %v", err)
	}

	l.Debug("repelling RLUSD (sum of loan interest) from System Account to owner/borrower")
	err = t.bc.PaymentRLUSDFromSystemAccount(owner, loan.Principal.InexactFloat64()/10)
	if err != nil {
		// l.Warn("failed to payment RLUSD from system account", "error", err)
		l.Error("failed to payment RLUSD from system account", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to payment RLUSD from system account: %v", err)
	}

	l.Debug("repelling RLUSD (loan body) from System Account to creditor/lender")
	err = t.bc.PaymentRLUSDFromSystemAccount(creditor, loan.Principal.InexactFloat64())
	if err != nil {
		// l.Warn("failed to payment RLUSD from system account", "error", err)
		l.Error("failed to payment RLUSD from system account", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to payment RLUSD from system account: %v", err)
	}

	l.Debug("minting debt token")
	debtToken := NewDebtMPToken(tokenID, owner.ClassicAddress.String(), creditor.ClassicAddress.String())
	hash, issuanceID, err := t.bc.MPTokenIssuanceCreate(owner, debtToken)
	if err != nil {
		l.Error("failed to mint debt token", "hash", hash, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to mint debt token: %v", err)
	}
	loan.SetDebtTokenID(issuanceID)

	l = l.With("debt_token_id", issuanceID)
	l.Debug("creditor/lender authorizing debt token")
	err = t.bc.AuthorizeMPToken(creditor, issuanceID)
	if err != nil {
		l.Warn("failed to authorize debt token", "error", err)
	}

	l.Debug("transferring debt token to creditor")
	hash, err = t.bc.TransferMPToken(owner, issuanceID, creditor.ClassicAddress.String())
	if err != nil {
		l.Error("failed to transfer debt token", "hash", hash, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to transfer debt token: %v", err)
	}

	l.Debug("transferring warrant token to creditor")
	err = t.bc.AuthorizeMPToken(creditor, tokenID)
	if err != nil {
		l.Warn("failed to authorize warrant token", "error", err)
	}

	mptHash, err := t.bc.TransferMPToken(owner, tokenID, creditor.ClassicAddress.String())
	if err != nil {
		l.Error("failed to transfer token", "hash", hash, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to transfer token: %v", err)
	}

	l.Debug("creditor/lender sending payment of RLUSD to owner/borrower with loan term",
		"amount", LoanAmount,
		"interest_rate", LoanInterestRate,
		"period", LoanPeriod,
	)

	err = t.bc.PaymentRLUSD(creditor, owner, loan.Principal.InexactFloat64())
	if err != nil {
		// l.Warn("failed to payment RLUSD", "error", err)
		l.Error("failed to payment RLUSD", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to payment RLUSD: %v", err)
	}

	l.Debug("add loan to interests tracking")
	t.loans.AddLoan(tokenID, loan)

	return &tokenv1.TransferToCreditorResponse{
		Error: nil,
		Token: &tokenv1.Token{
			Id: req.GetDocumentHash(),
			Transaction: &typesv1.Transaction{
				Id:        mptHash,
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

	err = t.bc.AuthorizeMPToken(owner, req.GetTokenId())
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
	tokenID := req.GetTokenId()
	l := t.logger.With("method", "BuyoutFromCreditorWithLoan",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
		"owner_address_id", req.GetOwnerAddressId(),
		"token_id", tokenID,
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

	l.Debug("returning loan body to creditor/lender")
	loan, err := t.loans.GetLoan(tokenID)
	if err != nil {
		l.Error("failed to get loan", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get loan: %v", err)
	}
	err = t.bc.PaymentRLUSD(owner, creditor, loan.Principal.InexactFloat64())
	if err != nil {
		l.Error("failed to payment RLUSD", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to payment RLUSD: %v", err)
	}

	l.Debug("returning and burning debt token to owner/borrower")
	hash, err := t.bc.TransferMPToken(creditor, loan.DebtTokenID, owner.ClassicAddress.String())
	if err != nil {
		l.Error("failed to transfer token", "debt_token_id", loan.DebtTokenID, "hash", hash, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to transfer token: %v", err)
	}
	t.loans.RemoveLoan(tokenID)
	err = t.bc.MPTokenIssuanceDestroy(owner, loan.DebtTokenID)
	if err != nil {
		l.Error("failed to destroy debt token", "debt_token_id", loan.DebtTokenID, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to destroy debt token: %v", err)
	}

	l.Debug("returning warrant token to owner/borrower")
	hash, err = t.bc.TransferMPToken(creditor, tokenID, owner.ClassicAddress.String())
	if err != nil {
		l.Error("failed to transfer token", "hash", hash, "error", err)
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
	tokenID := req.GetTokenId()
	l := t.logger.With("method", "TransferFromOwnerToWarehouseWithLoan",
		"document_hash", req.GetDocumentHash(),
		"creditor_address_id", req.GetCreditorAddressId(),
		"token_id", tokenID,
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

	l.Debug("returning and burning debt token to owner/borrower")
	loan, err := t.loans.GetLoan(tokenID)
	if err != nil {
		l.Error("failed to get loan", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get loan: %v", err)
	}

	hash, err := t.bc.TransferMPToken(creditor, loan.DebtTokenID, loan.OwnerWallet.ClassicAddress.String())
	if err != nil {
		l.Error("failed to transfer token", "debt_token_id", loan.DebtTokenID, "hash", hash, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to transfer token: %v", err)
	}
	t.loans.RemoveLoan(tokenID)

	err = t.bc.MPTokenIssuanceDestroy(loan.OwnerWallet, loan.DebtTokenID)
	if err != nil {
		l.Error("failed to destroy debt token", "debt_token_id", loan.DebtTokenID, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to destroy debt token: %v", err)
	}

	l.Debug("returning warrant token to warehouse")
	issuerAddr, err := t.bc.GetIssuerAddressFromIssuanceID(tokenID)
	if err != nil {
		l.Error("failed to get issuer address", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get issuer address: %v", err)
	}

	hash, err = t.bc.TransferMPToken(creditor, tokenID, issuerAddr)
	if err != nil {
		l.Error("failed to transfer token", "hash", hash, "error", err)
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
