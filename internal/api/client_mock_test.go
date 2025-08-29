package api

import (
	"fmt"
	"strconv"

	"github.com/CreatureDev/xrpl-go/client"
	"github.com/CreatureDev/xrpl-go/model/client/account"
	"github.com/CreatureDev/xrpl-go/model/client/server"
	"github.com/CreatureDev/xrpl-go/model/client/transactions"
	"github.com/CreatureDev/xrpl-go/model/ledger"
	"github.com/CreatureDev/xrpl-go/model/transactions/types"
)

// MockClient is a mock implementation of the client.Client interface for testing purposes
type MockClient struct {
	// Mock responses
	mockAccountInfo    *account.AccountInfoResponse
	mockServerInfo     *server.ServerInfoResponse
	mockSubmitResponse *transactions.SubmitResponse
	mockTxResponse     *transactions.TxResponse

	// Mock errors
	mockAccountInfoError error
	mockServerInfoError  error
	mockSubmitError      error
	mockTxError          error

	// Mock address and faucet
	mockAddress string
	mockFaucet  string

	// Call tracking
	accountInfoCalls int
	serverInfoCalls  int
	submitCalls      int
	txCalls          int
}

// NewMockClient creates a new MockClient instance
func NewMockClient() *MockClient {
	return &MockClient{
		mockAddress: "rMockAddress123456789",
		mockFaucet:  "https://faucet.testnet.rippletest.net/accounts",
	}
}

// SendRequest mocks the SendRequest method
func (m *MockClient) SendRequest(req client.XRPLRequest) (client.XRPLResponse, error) {
	return nil, fmt.Errorf("SendRequest not implemented in mock")
}

// Address returns the mock address
func (m *MockClient) Address() string {
	return m.mockAddress
}

// Faucet returns the mock faucet URL
func (m *MockClient) Faucet() string {
	return m.mockFaucet
}

// SetMockAccountInfo sets the mock response for account info requests
func (m *MockClient) SetMockAccountInfo(response *account.AccountInfoResponse, err error) {
	m.mockAccountInfo = response
	m.mockAccountInfoError = err
}

// SetMockServerInfo sets the mock response for server info requests
func (m *MockClient) SetMockServerInfo(response *server.ServerInfoResponse, err error) {
	m.mockServerInfo = response
	m.mockServerInfoError = err
}

// SetMockSubmitResponse sets the mock response for submit requests
func (m *MockClient) SetMockSubmitResponse(response *transactions.SubmitResponse, err error) {
	m.mockSubmitResponse = response
	m.mockSubmitError = err
}

// SetMockTxResponse sets the mock response for transaction info requests
func (m *MockClient) SetMockTxResponse(response *transactions.TxResponse, err error) {
	m.mockTxResponse = response
	m.mockTxError = err
}

// GetCallCounts returns the number of times each method was called
func (m *MockClient) GetCallCounts() map[string]int {
	return map[string]int{
		"AccountInfo": m.accountInfoCalls,
		"ServerInfo":  m.serverInfoCalls,
		"Submit":      m.submitCalls,
		"Tx":          m.txCalls,
	}
}

// ResetCallCounts resets all call counters to zero
func (m *MockClient) ResetCallCounts() {
	m.accountInfoCalls = 0
	m.serverInfoCalls = 0
	m.submitCalls = 0
	m.txCalls = 0
}

// MockAccountImpl mocks the Account interface
type MockAccountImpl struct {
	client *MockClient
}

func (m *MockAccountImpl) AccountInfo(req *account.AccountInfoRequest) (*account.AccountInfoResponse, client.XRPLResponse, error) {
	m.client.accountInfoCalls++
	if m.client.mockAccountInfoError != nil {
		return nil, nil, m.client.mockAccountInfoError
	}
	return m.client.mockAccountInfo, nil, nil
}

// Implement remaining Account interface methods with stubs
func (m *MockAccountImpl) AccountChannels(req *account.AccountChannelsRequest) (*account.AccountChannelsResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (m *MockAccountImpl) AccountCurrencies(req *account.AccountCurrenciesRequest) (*account.AccountCurrenciesResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (m *MockAccountImpl) AccountLines(req *account.AccountLinesRequest) (*account.AccountLinesResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (m *MockAccountImpl) AccountNFTs(req *account.AccountNFTsRequest) (*account.AccountNFTsResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (m *MockAccountImpl) AccountObjects(req *account.AccountObjectsRequest) (*account.AccountObjectsResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (m *MockAccountImpl) AccountOffers(req *account.AccountOffersRequest) (*account.AccountOffersResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (m *MockAccountImpl) AccountTransactions(req *account.AccountTransactionsRequest) (*account.AccountTransactionsResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// MockServerImpl mocks the Server interface
type MockServerImpl struct {
	client *MockClient
}

func (m *MockServerImpl) ServerInfo(req *server.ServerInfoRequest) (*server.ServerInfoResponse, client.XRPLResponse, error) {
	m.client.serverInfoCalls++
	if m.client.mockServerInfoError != nil {
		return nil, nil, m.client.mockServerInfoError
	}
	return m.client.mockServerInfo, nil, nil
}

// Implement remaining Server interface methods with stubs
func (m *MockServerImpl) Fee(req *server.FeeRequest) (*server.FeeResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (m *MockServerImpl) Manifest(req *server.ManifestRequest) (*server.ManifestResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (m *MockServerImpl) ServerState(req *server.ServerStateRequest) (*server.ServerStateResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// MockTransactionImpl mocks the Transaction interface
type MockTransactionImpl struct {
	client *MockClient
}

func (m *MockTransactionImpl) Submit(req *transactions.SubmitRequest) (*transactions.SubmitResponse, client.XRPLResponse, error) {
	m.client.submitCalls++
	if m.client.mockSubmitError != nil {
		return nil, nil, m.client.mockSubmitError
	}
	return m.client.mockSubmitResponse, nil, nil
}

func (m *MockTransactionImpl) Tx(req *transactions.TxRequest) (*transactions.TxResponse, client.XRPLResponse, error) {
	m.client.txCalls++
	if m.client.mockTxError != nil {
		return nil, nil, m.client.mockTxError
	}
	return m.client.mockTxResponse, nil, nil
}

// Implement remaining Transaction interface methods with stubs
func (m *MockTransactionImpl) SubmitMultisigned(req *transactions.SubmitMultisignedRequest) (*transactions.SubmitMultisignedResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (m *MockTransactionImpl) TransactionEntry(req *transactions.TransactionEntryRequest) (*transactions.TransactionEntryResponse, client.XRPLResponse, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

// CreateMockXRPLClient creates a client.XRPLClient with mocked implementations
func CreateMockXRPLClient(mockClient *MockClient) *client.XRPLClient {
	xrplClient := client.NewXRPLClient(mockClient)

	// Override the Server implementation with our mock
	xrplClient.Server = &MockServerImpl{client: mockClient}

	return xrplClient
}

// Helper functions to create mock responses

// CreateMockAccountInfoResponse creates a mock account info response
func CreateMockAccountInfoResponse(address string, sequence uint32, balance string) *account.AccountInfoResponse {
	balanceInt, _ := strconv.ParseUint(balance, 10, 64)
	return &account.AccountInfoResponse{
		AccountData: ledger.AccountRoot{
			Account:  types.Address(address),
			Sequence: sequence,
			Balance:  types.XRPCurrencyAmount(balanceInt),
		},
	}
}

// CreateMockServerInfoResponse creates a mock server info response
func CreateMockServerInfoResponse(baseFeeXRP float64) *server.ServerInfoResponse {
	return &server.ServerInfoResponse{
		Info: server.ServerInfo{
			ValidatedLedger: &server.ServerLedgerInfo{
				BaseFeeXRP: float32(baseFeeXRP),
			},
		},
	}
}

// CreateMockSubmitResponse creates a mock submit response
func CreateMockSubmitResponse(engineResult string, engineResultCode int, engineResultMessage string) *transactions.SubmitResponse {
	return &transactions.SubmitResponse{
		EngineResult:        engineResult,
		EngineResultCode:    engineResultCode,
		EngineResultMessage: engineResultMessage,
	}
}

// CreateMockTxResponse creates a mock transaction response
func CreateMockTxResponse() *transactions.TxResponse {
	return &transactions.TxResponse{}
}
