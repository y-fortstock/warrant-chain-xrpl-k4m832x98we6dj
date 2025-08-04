package api

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	accountv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/account/v1"
	typesv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/types/v1"
)

// Account is an implementation of accountv1.AccountAPIServer.
type Account struct {
	accountv1.UnimplementedAccountAPIServer
	bc     *Blockchain
	logger *slog.Logger
}

// NewAccount returns a new Account implementation.
func NewAccount(l *slog.Logger, bc *Blockchain) *Account {
	return &Account{logger: l, bc: bc}
}

// Create creates a new ETH account with a password.
func (a *Account) Create(ctx context.Context, req *accountv1.CreateRequest) (*accountv1.CreateResponse, error) {
	a.logger.Info("create account")
	seeds := strings.Split(req.Password, "-")
	address, _, _, err := a.bc.GetXRPLWallet(seeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", seeds[1]))
	if err != nil {
		a.logger.Error("failed to get XRPL address", "error", err)
		return nil, err
	}

	a.logger.Info("account created", "address", address)
	return &accountv1.CreateResponse{
		Account: &accountv1.Account{
			Id: address,
		},
	}, nil
}

// Deposit deposits XRP in drops from system account.
func (a *Account) Deposit(ctx context.Context, req *accountv1.DepositRequest) (*accountv1.DepositResponse, error) {
	a.logger.Info("deposit request", "account", req.AccountId, "amount", req.WeiAmount)

	info, err := a.bc.GetAccountInfo(a.bc.systemAccount)
	if err != nil {
		a.logger.Error("failed to get system account balance", "error", err)
		return nil, err
	}
	balance := uint64(info.AccountData.Balance)
	sequence := info.AccountData.Sequence

	feeRaw, reserveRaw, err := a.bc.GetBaseFeeAndReserve()
	if err != nil {
		a.logger.Error("failed to get base fee and reserve", "error", err)
		return nil, err
	}
	fee := uint64(feeRaw * xrpToDrops * 120 / 100) // 20% margin
	reserve := uint64(reserveRaw * xrpToDrops)

	dropsToTransfer, err := strconv.ParseUint(req.WeiAmount, 10, 64)
	if err != nil {
		a.logger.Error("failed to parse wei amount", "error", err, "weiAmount", req.WeiAmount)
		return nil, fmt.Errorf("invalid wei amount: %s", req.WeiAmount)
	}

	if balance < dropsToTransfer+fee+reserve {
		a.logger.Error("system account balance is less than fee + reserve + drops to transfer",
			"balance", balance,
			"dropsToTransfer", dropsToTransfer,
			"fee", fee,
			"reserve", reserve)
		return nil, fmt.Errorf("system account balance is less than fee + reserve + drops to transfer: %d < %d",
			balance, dropsToTransfer+fee+reserve)
	}

	a.logger.Info("payment from system account", "account", req.AccountId, "fee", fee, "dropsToTransfer", dropsToTransfer, "sequence", sequence)
	txHash, err := a.bc.PaymentFromSystemAccount(req.AccountId, fee, dropsToTransfer, sequence)
	if err != nil {
		a.logger.Error("failed to payment from system account",
			"error", err,
			"account", req.AccountId,
			"fee", fee,
			"dropsToTransfer", dropsToTransfer,
			"sequence", sequence)
		return nil, err
	}

	a.logger.Info("deposit response", "account", req.AccountId, "txHash", txHash)
	return &accountv1.DepositResponse{
		Transaction: &typesv1.Transaction{
			Id:          txHash,
			BlockNumber: []byte(strconv.FormatUint(uint64(sequence), 10)),
			BlockTime:   uint64(time.Now().Unix()),
		},
	}, nil
}

// ClearBalance clears the account balance.
func (a *Account) ClearBalance(ctx context.Context, req *accountv1.ClearBalanceRequest) (*accountv1.ClearBalanceResponse, error) {
	a.logger.Info("clear balance request", "account", req.AccountId)

	seeds := strings.Split(req.AccountPassword, "-")
	address, public, private, err := a.bc.GetXRPLWallet(seeds[0], fmt.Sprintf("m/44'/144'/0'/0/%s", seeds[1]))
	if err != nil {
		a.logger.Error("failed to get XRPL address", "error", err)
		return nil, err
	}
	if address != req.AccountId {
		a.logger.Error("account id mismatch", "address", address, "accountId", req.AccountId)
		return nil, fmt.Errorf("account id mismatch: %s != %s", address, req.AccountId)
	}

	info, err := a.bc.GetAccountInfo(req.AccountId)
	if err != nil {
		a.logger.Error("failed to get account balance", "error", err, "account", req.AccountId)
		return nil, err
	}
	balance := uint64(info.AccountData.Balance)
	sequence := info.AccountData.Sequence

	feeRaw, reserveRaw, err := a.bc.GetBaseFeeAndReserve()
	if err != nil {
		a.logger.Error("failed to get base fee and reserve", "error", err)
		return nil, err
	}
	fee := uint64(feeRaw * xrpToDrops * 120 / 100) // 20% margin
	reserve := uint64(reserveRaw * xrpToDrops)

	feeWithReserve := fee + reserve
	if balance <= feeWithReserve {
		a.logger.Warn("account balance is less or equal than fee + reserve", "balance", balance, "fee", feeWithReserve)
		return nil, fmt.Errorf("account balance is less or equal than fee + reserve: %d <= %d", balance, feeWithReserve)
	}
	amount := balance - feeWithReserve

	a.logger.Info("payment to system account",
		"account", req.AccountId,
		"fee", fee,
		"reserve", reserve,
		"amount", amount,
		"sequence", sequence)
	txHash, err := a.bc.PaymentToSystemAccount(address, public, private, fee, amount, sequence)
	if err != nil {
		a.logger.Error("failed to payment to system account",
			"error", err,
			"account", req.AccountId,
			"fee", fee,
			"reserve", reserve,
			"amount", amount,
			"sequence", sequence)
		return nil, err
	}

	return &accountv1.ClearBalanceResponse{
		Transaction: &typesv1.Transaction{
			Id:          txHash,
			BlockNumber: []byte(strconv.FormatUint(uint64(sequence), 10)),
			BlockTime:   uint64(time.Now().Unix()),
		},
	}, nil
}

// GetBalance gets the account balance.
func (a *Account) GetBalance(ctx context.Context, req *accountv1.GetBalanceRequest) (*accountv1.GetBalanceResponse, error) {
	a.logger.Info("get balance request", "account", req.AccountId)

	info, err := a.bc.GetAccountInfo(req.AccountId)
	if err != nil {
		a.logger.Error("failed to get account balance", "error", err, "account", req.AccountId)
		return nil, err
	}
	balance := uint64(info.AccountData.Balance)

	a.logger.Info("account balance response", "account", req.AccountId, "balance", balance)
	return &accountv1.GetBalanceResponse{
		Balance: strconv.FormatUint(balance, 10),
	}, nil
}
