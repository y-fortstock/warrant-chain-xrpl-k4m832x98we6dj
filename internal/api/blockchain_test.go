package api

import (
	"fmt"
	"testing"

	"github.com/CreatureDev/xrpl-go/model/client/account"
	clientcommon "github.com/CreatureDev/xrpl-go/model/client/common"
	"github.com/CreatureDev/xrpl-go/model/client/server"
	"github.com/CreatureDev/xrpl-go/model/ledger"
	"github.com/CreatureDev/xrpl-go/model/transactions/types"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/config"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"
)

// XRPLClientInterface определяет интерфейс для XRPL клиента
type XRPLClientInterface interface {
	Account() AccountInterface
	Server() ServerInterface
}

// AccountInterface определяет интерфейс для работы с аккаунтами
type AccountInterface interface {
	AccountInfo(req *account.AccountInfoRequest) (*account.AccountInfoResponse, interface{}, error)
}

// ServerInterface определяет интерфейс для работы с сервером
type ServerInterface interface {
	Fee(req *server.FeeRequest) (*server.FeeResponse, interface{}, error)
}

// BlockchainInterface определяет интерфейс для Blockchain
type BlockchainInterface interface {
	GetXRPLWallet(hexSeed string, path string) (address string, private string, err error)
	GetAccountBalance(address string) (uint64, error)
	GetBaseFee() (uint64, error)
}

// TestBlockchain - тестовая версия Blockchain с мок клиентом
type TestBlockchain struct {
	accountMock   *MockAccount
	serverMock    *MockServer
	systemAccount string
	systemSecret  string
}

func (tb *TestBlockchain) GetXRPLWallet(hexSeed string, path string) (address string, private string, err error) {
	// Используем реальную логику для GetXRPLWallet, так как она не зависит от XRPL клиента
	extendedKey, err := crypto.GetExtendedKeyFromHexSeedWithPath(hexSeed, path)
	if err != nil {
		return "", "", fmt.Errorf("failed to get extended key from hex seed: %w", err)
	}
	address, private, err = crypto.GetXRPLWallet(extendedKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to get xrpl wallet: %w", err)
	}
	return address, private, nil
}

func (tb *TestBlockchain) GetAccountBalance(address string) (uint64, error) {
	xrplReq := &account.AccountInfoRequest{
		Account:     types.Address(address),
		LedgerIndex: clientcommon.VALIDATED,
	}
	resp, xrplRes, err := tb.accountMock.AccountInfo(xrplReq)
	if err != nil {
		return 0, fmt.Errorf("failed to get account info for %s: %w (xrplRes: %v)", address, err, xrplRes)
	}

	return uint64(resp.AccountData.Balance), nil
}

func (tb *TestBlockchain) GetBaseFee() (uint64, error) {
	xrplReq := &server.FeeRequest{}
	resp, xrplRes, err := tb.serverMock.Fee(xrplReq)
	if err != nil {
		return 0, fmt.Errorf("failed to get base fee: %w (xrplRes: %v)", err, xrplRes)
	}
	return uint64(resp.Drops.BaseFee), nil
}

// MockXRPLClient - мок для XRPL клиента
type MockXRPLClient struct {
	accountMock *MockAccount
	serverMock  *MockServer
}

func (m *MockXRPLClient) Account() AccountInterface {
	return m.accountMock
}

func (m *MockXRPLClient) Server() ServerInterface {
	return m.serverMock
}

// MockAccount - мок для работы с аккаунтами
type MockAccount struct {
	accountInfoFunc func(req *account.AccountInfoRequest) (*account.AccountInfoResponse, interface{}, error)
}

func (m *MockAccount) AccountInfo(req *account.AccountInfoRequest) (*account.AccountInfoResponse, interface{}, error) {
	if m.accountInfoFunc != nil {
		return m.accountInfoFunc(req)
	}
	return nil, nil, nil
}

// MockServer - мок для работы с сервером
type MockServer struct {
	feeFunc func(req *server.FeeRequest) (*server.FeeResponse, interface{}, error)
}

func (m *MockServer) Fee(req *server.FeeRequest) (*server.FeeResponse, interface{}, error) {
	if m.feeFunc != nil {
		return m.feeFunc(req)
	}
	return nil, nil, nil
}

// createMockBlockchain создает blockchain с мок клиентом для тестирования
func createMockBlockchain() *TestBlockchain {
	return &TestBlockchain{
		accountMock:   &MockAccount{},
		serverMock:    &MockServer{},
		systemAccount: "rTestAccount",
		systemSecret:  "testSecret",
	}
}

var (
	validHexSeed   = "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"
	validAddress   = "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC"
	invalidHexSeed = "invalid_hex_seed"
)

func TestNewBlockchain(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.NetworkConfig
		wantErr bool
	}{
		{
			name: "valid network config",
			cfg: config.NetworkConfig{
				URL: "https://s.altnet.rippletest.net:51234",
			},
			wantErr: false,
		},
		{
			name: "invalid URL",
			cfg: config.NetworkConfig{
				URL: "invalid://url",
			},
			wantErr: false,
		},
		{
			name: "empty URL",
			cfg: config.NetworkConfig{
				URL: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockchain, err := NewBlockchain(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBlockchain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && blockchain == nil {
				t.Error("NewBlockchain() returned nil blockchain when no error expected")
			}
			if !tt.wantErr && blockchain.xrplClient == nil {
				t.Error("NewBlockchain() returned blockchain with nil xrplClient")
			}
		})
	}
}

func TestBlockchain_GetXRPLWallet(t *testing.T) {
	// Создаем мок blockchain
	blockchain := createMockBlockchain()

	tests := []struct {
		name        string
		hexSeed     string
		path        string
		wantErr     bool
		expectEmpty bool
	}{
		{
			name:        "valid hex seed and path",
			hexSeed:     validHexSeed,
			path:        "44'/144'/0'/0/0",
			wantErr:     false,
			expectEmpty: false,
		},
		{
			name:        "valid hex seed with different path",
			hexSeed:     validHexSeed,
			path:        "44'/144'/0'/0/1",
			wantErr:     false,
			expectEmpty: false,
		},
		{
			name:        "invalid hex seed",
			hexSeed:     invalidHexSeed,
			path:        "44'/144'/0'/0/0",
			wantErr:     true,
			expectEmpty: true,
		},
		{
			name:        "empty hex seed",
			hexSeed:     "",
			path:        "44'/144'/0'/0/0",
			wantErr:     true,
			expectEmpty: true,
		},
		{
			name:        "invalid path",
			hexSeed:     validHexSeed,
			path:        "invalid/path",
			wantErr:     true,
			expectEmpty: true,
		},
		{
			name:        "empty path",
			hexSeed:     validHexSeed,
			path:        "",
			wantErr:     true,
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, private, err := blockchain.GetXRPLWallet(tt.hexSeed, tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetXRPLWallet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectEmpty {
				if address != "" {
					t.Errorf("GetXRPLWallet() address = %v, want empty string", address)
				}
				if private != "" {
					t.Errorf("GetXRPLWallet() private = %v, want empty string", private)
				}
			} else {
				if address == "" {
					t.Error("GetXRPLWallet() returned empty address when no error expected")
				}
				if private == "" {
					t.Error("GetXRPLWallet() returned empty private key when no error expected")
				}
				// Проверяем, что адрес начинается с 'r' (XRPL адрес)
				if len(address) > 0 && address[0] != 'r' {
					t.Errorf("GetXRPLWallet() address = %v, should start with 'r'", address)
				}
			}
		})
	}
}

func TestBlockchain_GetAccountBalance(t *testing.T) {
	// Создаем мок blockchain
	blockchain := createMockBlockchain()

	tests := []struct {
		name        string
		address     string
		mockBalance uint64
		mockError   error
		wantErr     bool
		wantBalance uint64
	}{
		{
			name:        "valid address with balance",
			address:     validAddress,
			mockBalance: 1000000,
			mockError:   nil,
			wantErr:     false,
			wantBalance: 1000000,
		},
		{
			name:        "valid address with zero balance",
			address:     validAddress,
			mockBalance: 0,
			mockError:   nil,
			wantErr:     false,
			wantBalance: 0,
		},
		{
			name:        "network error",
			address:     validAddress,
			mockBalance: 0,
			mockError:   fmt.Errorf("network error"),
			wantErr:     true,
			wantBalance: 0,
		},
		{
			name:        "empty address",
			address:     "",
			mockBalance: 0,
			mockError:   fmt.Errorf("invalid address"),
			wantErr:     true,
			wantBalance: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок для этого теста
			blockchain.accountMock.accountInfoFunc = func(req *account.AccountInfoRequest) (*account.AccountInfoResponse, interface{}, error) {
				if tt.mockError != nil {
					return nil, nil, tt.mockError
				}
				return &account.AccountInfoResponse{
					AccountData: ledger.AccountRoot{
						Balance: types.XRPCurrencyAmount(tt.mockBalance),
					},
				}, nil, nil
			}

			balance, err := blockchain.GetAccountBalance(tt.address)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAccountBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if balance != tt.wantBalance {
					t.Errorf("GetAccountBalance() balance = %v, want %v", balance, tt.wantBalance)
				}
			}
		})
	}
}

func TestBlockchain_GetBaseFee(t *testing.T) {
	// Создаем мок blockchain
	blockchain := createMockBlockchain()

	tests := []struct {
		name      string
		mockFee   uint64
		mockError error
		wantErr   bool
		wantFee   uint64
	}{
		{
			name:      "valid base fee",
			mockFee:   10000,
			mockError: nil,
			wantErr:   false,
			wantFee:   10000,
		},
		{
			name:      "zero base fee",
			mockFee:   0,
			mockError: nil,
			wantErr:   false,
			wantFee:   0,
		},
		{
			name:      "network error",
			mockFee:   0,
			mockError: fmt.Errorf("network error"),
			wantErr:   true,
			wantFee:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок для этого теста
			blockchain.serverMock.feeFunc = func(req *server.FeeRequest) (*server.FeeResponse, interface{}, error) {
				if tt.mockError != nil {
					return nil, nil, tt.mockError
				}
				return &server.FeeResponse{
					Drops: server.FeeDrops{
						BaseFee: types.XRPCurrencyAmount(tt.mockFee),
					},
				}, nil, nil
			}

			fee, err := blockchain.GetBaseFee()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBaseFee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if fee != tt.wantFee {
					t.Errorf("GetBaseFee() fee = %v, want %v", fee, tt.wantFee)
				}
			}
		})
	}
}
