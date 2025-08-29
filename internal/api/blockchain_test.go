package api

import (
	"errors"
	"testing"

	"github.com/CreatureDev/xrpl-go/model/client/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockchain_GetBaseFeeAndReserve(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  *server.ServerInfoResponse
		mockError     error
		expectedInfo  *server.ServerLedgerInfo
		expectedError string
		expectedCalls int
	}{
		{
			name: "successful response",
			mockResponse: &server.ServerInfoResponse{
				Info: server.ServerInfo{
					ValidatedLedger: &server.ServerLedgerInfo{
						BaseFeeXRP:     0.00001,
						ReserveBaseXRP: 10.0,
						ReserveIncXRP:  2.0,
					},
				},
			},
			mockError: nil,
			expectedInfo: &server.ServerLedgerInfo{
				BaseFeeXRP:     0.00001,
				ReserveBaseXRP: 10.0,
				ReserveIncXRP:  2.0,
			},
			expectedError: "",
			expectedCalls: 1,
		},
		{
			name:          "server error",
			mockResponse:  nil,
			mockError:     errors.New("network error"),
			expectedInfo:  nil,
			expectedError: "failed to get server info: network error",
			expectedCalls: 1,
		},
		{
			name: "nil validated ledger",
			mockResponse: &server.ServerInfoResponse{
				Info: server.ServerInfo{
					ValidatedLedger: nil,
				},
			},
			mockError:     nil,
			expectedInfo:  nil,
			expectedError: "",
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := NewMockClient()
			mockClient.SetMockServerInfo(tt.mockResponse, tt.mockError)

			// Create blockchain instance with mock client
			blockchain := &Blockchain{
				xrplClient: CreateMockXRPLClient(mockClient),
			}

			// Call the method under test
			result, err := blockchain.GetBaseFeeAndReserve()

			// Verify call count
			callCounts := mockClient.GetCallCounts()
			require.Equal(t, tt.expectedCalls, callCounts["ServerInfo"], "ServerInfo should be called the expected number of times")

			// Verify results
			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				if tt.expectedInfo != nil {
					require.NotNil(t, result)
					assert.Equal(t, tt.expectedInfo.BaseFeeXRP, result.BaseFeeXRP)
					assert.Equal(t, tt.expectedInfo.ReserveBaseXRP, result.ReserveBaseXRP)
					assert.Equal(t, tt.expectedInfo.ReserveIncXRP, result.ReserveIncXRP)
				} else {
					require.Nil(t, result)
				}
			}
		})
	}
}

func TestBlockchain_GetBaseFeeAndReserve_Integration(t *testing.T) {
	// Test with realistic mock data
	mockClient := NewMockClient()

	// Set up a realistic server info response
	mockResponse := CreateMockServerInfoResponse(0.00001)
	mockClient.SetMockServerInfo(mockResponse, nil)

	blockchain := &Blockchain{
		xrplClient: CreateMockXRPLClient(mockClient),
	}

	// Test the method
	result, err := blockchain.GetBaseFeeAndReserve()

	// Verify no error
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify the response contains expected data
	assert.Equal(t, float32(0.00001), result.BaseFeeXRP)

	// Verify call tracking
	callCounts := mockClient.GetCallCounts()
	require.Equal(t, 1, callCounts["ServerInfo"], "ServerInfo should be called once")

	// Test call count reset
	mockClient.ResetCallCounts()
	callCounts = mockClient.GetCallCounts()
	require.Equal(t, 0, callCounts["ServerInfo"], "Call count should be reset to zero")
}
