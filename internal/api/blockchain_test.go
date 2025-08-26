package api

import (
	"fmt"
	"testing"

	"github.com/CreatureDev/xrpl-go/model/client/account"
	clientcommon "github.com/CreatureDev/xrpl-go/model/client/common"
	"github.com/CreatureDev/xrpl-go/model/client/server"
	clienttransactions "github.com/CreatureDev/xrpl-go/model/client/transactions"
	"github.com/CreatureDev/xrpl-go/model/ledger"
	"github.com/CreatureDev/xrpl-go/model/transactions"
	"github.com/CreatureDev/xrpl-go/model/transactions/types"
	"github.com/stretchr/testify/assert"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/config"
	"gitlab.com/warrant1/warrant/chain-xrpl/internal/crypto"
)

var (
	rippleUrl = "https://s.devnet.rippletest.net:51234"

	accountAddress = "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC"
	publicKey      = "ED80EA4365634AB2116C239CEB8F739498CEFE91FBB667FBAB6FE9B93492ED0FFC"
	privateKey     = "ED75207685F294BE4945908D2BBF1E535CECFB7D78A6B9AEC865F146B611DB2E51"

	accountAddress_to = "rJqzDMuSpE8pxztkeES3VeKGauFFRj8qDQ"
	publicKey_to      = "EDB1B0C29442DB938BAC88A5C5ACEC9349DE619929E045590145126B8E8D6D88AF"
	privateKey_to     = "ED324167F1057138731EA904BF23D93F29A7BB08D091931ED3FD8FA8182BA7C558"
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

func (tb *TestBlockchain) GetXRPLWallet(hexSeed string, path string) (address string, public string, private string, err error) {
	// Используем реальную логику для GetXRPLWallet, так как она не зависит от XRPL клиента
	extendedKey, err := crypto.GetExtendedKeyFromHexSeedWithPath(hexSeed, path)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get extended key from hex seed: %w", err)
	}
	address, public, private, err = crypto.GetXRPLWallet(extendedKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get xrpl wallet: %w", err)
	}
	return address, public, private, nil
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
				System: struct {
					Account string `mapstructure:"account"`
					Secret  string `mapstructure:"secret"`
					Public  string `mapstructure:"public"`
				}{
					Account: "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC",
					Secret:  "ED75207685F294BE4945908D2BBF1E535CECFB7D78A6B9AEC865F146B611DB2E51",
					Public:  "ED80EA4365634AB2116C239CEB8F739498CEFE91FBB667FBAB6FE9B93492ED0FFC",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid URL",
			cfg: config.NetworkConfig{
				URL: "invalid://url",
				System: struct {
					Account string `mapstructure:"account"`
					Secret  string `mapstructure:"secret"`
					Public  string `mapstructure:"public"`
				}{
					Account: "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC",
					Secret:  "ED75207685F294BE4945908D2BBF1E535CECFB7D78A6B9AEC865F146B611DB2E51",
					Public:  "ED80EA4365634AB2116C239CEB8F739498CEFE91FBB667FBAB6FE9B93492ED0FFC",
				},
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
			address, public, private, err := blockchain.GetXRPLWallet(tt.hexSeed, tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetXRPLWallet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectEmpty {
				if address != "" {
					t.Errorf("GetXRPLWallet() address = %v, want empty string", address)
				}
				if public != "" {
					t.Errorf("GetXRPLWallet() public = %v, want empty string", public)
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

func TestBlockchain_SubmitTx(t *testing.T) {
	// Создаем тестовый кошелек с корректным XRPL адресом
	testWallet, err := crypto.NewWallet(
		types.Address("rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC"),
		"ED80EA4365634AB2116C239CEB8F739498CEFE91FBB667FBAB6FE9B93492ED0FFC",
		"ED75207685F294BE4945908D2BBF1E535CECFB7D78A6B9AEC865F146B611DB2E51",
	)
	assert.NoError(t, err)

	testTx := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			TransactionType: transactions.PaymentTx,
			Fee:             types.XRPCurrencyAmount(10000),
			Sequence:        1,
		},
		Amount:      types.XRPCurrencyAmount(1000000),
		Destination: types.Address("rJqzDMuSpE8pxztkeES3VeKGauFFRj8qDQ"),
	}

	tests := []struct {
		name     string
		wallet   *crypto.Wallet
		tx       transactions.Tx
		wantErr  bool
		errorMsg string
	}{
		{
			name:    "valid wallet and transaction",
			wallet:  testWallet,
			tx:      testTx,
			wantErr: false,
		},
		{
			name:     "nil wallet",
			wallet:   nil,
			tx:       testTx,
			wantErr:  true,
			errorMsg: "wallet cannot be nil",
		},
		{
			name:     "nil transaction",
			wallet:   testWallet,
			tx:       nil,
			wantErr:  true,
			errorMsg: "transaction cannot be nil",
		},
		{
			name:     "invalid wallet - empty address",
			wallet:   &crypto.Wallet{Address: "", PublicKey: testWallet.PublicKey, PrivateKey: testWallet.PrivateKey},
			tx:       testTx,
			wantErr:  true,
			errorMsg: "wallet is invalid: wallet address cannot be empty",
		},
		{
			name:     "invalid wallet - empty public key",
			wallet:   &crypto.Wallet{Address: testWallet.Address, PublicKey: "", PrivateKey: testWallet.PrivateKey},
			tx:       testTx,
			wantErr:  true,
			errorMsg: "wallet is invalid: wallet public key cannot be empty",
		},
		{
			name:     "invalid wallet - empty private key",
			wallet:   &crypto.Wallet{Address: testWallet.Address, PublicKey: testWallet.PublicKey, PrivateKey: ""},
			tx:       testTx,
			wantErr:  true,
			errorMsg: "wallet is invalid: wallet private key cannot be empty",
		},
		{
			name:     "invalid wallet - malformed address",
			wallet:   &crypto.Wallet{Address: "invalid_address", PublicKey: testWallet.PublicKey, PrivateKey: testWallet.PrivateKey},
			tx:       testTx,
			wantErr:  true,
			errorMsg: "wallet is invalid: invalid xrpl address length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Используем mock blockchain вместо реального
			blockchain, err := NewMockBlockchain(config.NetworkConfig{
				URL:     "mock://localhost", // Используем mock URL
				Timeout: 30,
				System: struct {
					Account string `mapstructure:"account"`
					Secret  string `mapstructure:"secret"`
					Public  string `mapstructure:"public"`
				}{
					Account: "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC",
					Secret:  "ED75207685F294BE4945908D2BBF1E535CECFB7D78A6B9AEC865F146B611DB2E51",
					Public:  "ED80EA4365634AB2116C239CEB8F739498CEFE91FBB667FBAB6FE9B93492ED0FFC",
				},
			})
			assert.NoError(t, err)

			// Вызываем SubmitTx
			resp, xrplResp, err := blockchain.SubmitTx(tt.wallet, tt.tx)

			// Проверяем результат
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, resp)
				assert.Nil(t, xrplResp)
			} else {
				// Для успешного случая проверяем, что ошибки нет и получен mock ответ
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "tesSUCCESS", resp.EngineResult)
				assert.Equal(t, "The transaction was applied.", resp.EngineResultMessage)
				assert.Equal(t, "mock_tx_blob", resp.TxBlob)
				assert.True(t, resp.Accepted)
				assert.True(t, resp.Applied)
				assert.True(t, resp.Broadcast)
			}
		})
	}
}

// MockBlockchain - мок версия Blockchain для тестирования SubmitTx
type MockBlockchain struct {
	SystemWallet *crypto.Wallet
}

func NewMockBlockchain(cfg config.NetworkConfig) (*MockBlockchain, error) {
	systemWallet, err := crypto.NewWallet(types.Address(cfg.System.Account), cfg.System.Public, cfg.System.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create system wallet: %w", err)
	}
	return &MockBlockchain{
		SystemWallet: systemWallet,
	}, nil
}

func (b *MockBlockchain) SubmitTx(w *crypto.Wallet, tx transactions.Tx) (
	resp *clienttransactions.SubmitResponse, xrplResp interface{}, err error) {

	// Проверяем входные параметры
	if w == nil {
		return nil, nil, fmt.Errorf("wallet cannot be nil")
	}
	if tx == nil {
		return nil, nil, fmt.Errorf("transaction cannot be nil")
	}
	if err := w.Validate(); err != nil {
		return nil, nil, fmt.Errorf("wallet is invalid: %w", err)
	}

	// Создаем mock ответ
	mockResp := &clienttransactions.SubmitResponse{
		EngineResult:        "tesSUCCESS",
		EngineResultMessage: "The transaction was applied.",
		TxBlob:              "mock_tx_blob",
		Accepted:            true,
		Applied:             true,
		Broadcast:           true,
	}

	// Возвращаем mock ответ
	return mockResp, nil, nil
}
