package api

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	accountv1 "gitlab.com/warrant1/warrant/protobuf/blockchain/account/v1"
)

// createTestAccount creates a test instance of Account API
func createTestAccount() *Account {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// Create nil blockchain since the Create method doesn't use it
	return NewAccount(logger, nil)
}

var (
	testHexSeed = "434670347c6bb7c791e3629fc79c38307315d625fc5b448a601abda6ba54f7efd0cfe70bf769f7e3545c970851f6fe9132ad658101ed1ff9cb2edfeb5dd2d19f"
	testAddress = "rKxt8PgUy4ggMY53GXuqU6i2aJ2HymW2YC"
)

func TestAccount_Create(t *testing.T) {
	accountAPI := createTestAccount()
	ctx := context.Background()

	tests := []struct {
		name        string
		password    string
		wantErr     bool
		errorMsg    string
		expectEmpty bool
	}{
		{
			name:        "valid hex seed with derivation index 0",
			password:    testHexSeed + "-0",
			wantErr:     false,
			expectEmpty: false,
		},
		{
			name:        "valid hex seed with derivation index 1",
			password:    testHexSeed + "-1",
			wantErr:     false,
			expectEmpty: false,
		},
		{
			name:        "valid hex seed with derivation index 10",
			password:    testHexSeed + "-10",
			wantErr:     false,
			expectEmpty: false,
		},
		{
			name:        "empty password",
			password:    "",
			wantErr:     true,
			expectEmpty: true,
		},
		{
			name:        "password without dash separator",
			password:    testHexSeed,
			wantErr:     true,
			expectEmpty: true,
		},
		{
			name:        "password with multiple dashes",
			password:    testHexSeed + "-0-1",
			wantErr:     true, // Code now validates that split gives exactly 2 elements
			expectEmpty: true,
		},
		{
			name:        "invalid hex seed",
			password:    "invalid_hex_seed-0",
			wantErr:     true,
			expectEmpty: true,
		},
		{
			name:        "invalid derivation index - not a number",
			password:    testHexSeed + "-abc",
			wantErr:     true,
			expectEmpty: true,
		},
		{
			name:        "very large derivation index",
			password:    testHexSeed + "-999999",
			wantErr:     false,
			expectEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &accountv1.CreateRequest{
				Password: tt.password,
			}

			resp, err := accountAPI.Create(ctx, req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Account)

				if tt.expectEmpty {
					assert.Empty(t, resp.Account.Id)
				} else {
					assert.NotEmpty(t, resp.Account.Id)
					// Check that address starts with 'r' (XRPL address)
					assert.Equal(t, uint8('r'), resp.Account.Id[0])
					// Check that address has correct length (25-34 characters for XRPL)
					assert.GreaterOrEqual(t, len(resp.Account.Id), 25)
					assert.LessOrEqual(t, len(resp.Account.Id), 34)
				}
			}
		})
	}
}

func TestAccount_Create_EdgeCases(t *testing.T) {
	accountAPI := createTestAccount()
	ctx := context.Background()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "password with leading dash",
			password: "-" + testHexSeed + "-0",
			wantErr:  true,
		},
		{
			name:     "password with trailing dash",
			password: testHexSeed + "-0-",
			wantErr:  true, // Code now validates that split gives exactly 2 elements
		},
		{
			name:     "password with only dash",
			password: "-",
			wantErr:  true,
		},
		{
			name:     "password with empty hex seed",
			password: "-0",
			wantErr:  true,
		},
		{
			name:     "password with empty derivation index",
			password: testHexSeed + "-",
			wantErr:  true,
		},
		{
			name:     "password with spaces",
			password: testHexSeed + " - 0",
			wantErr:  true,
		},
		{
			name:     "password with tabs",
			password: testHexSeed + "\t-0",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &accountv1.CreateRequest{
				Password: tt.password,
			}

			resp, err := accountAPI.Create(ctx, req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestAccount_Create_Consistency(t *testing.T) {
	accountAPI := createTestAccount()
	ctx := context.Background()

	// Test consistency: the same password should always give the same address
	password := testHexSeed + "-0"
	req := &accountv1.CreateRequest{
		Password: password,
	}

	resp1, err1 := accountAPI.Create(ctx, req)
	assert.NoError(t, err1)
	assert.NotNil(t, resp1)

	resp2, err2 := accountAPI.Create(ctx, req)
	assert.NoError(t, err2)
	assert.NotNil(t, resp2)

	// Addresses should be the same
	assert.Equal(t, resp1.Account.Id, resp2.Account.Id)
}

func TestAccount_Create_DifferentDerivationPaths(t *testing.T) {
	accountAPI := createTestAccount()
	ctx := context.Background()

	// Test that different derivation indices give different addresses
	indices := []string{"0", "1", "2", "10", "100", "1000", "10000", "100000"}
	addresses := make(map[string]bool)

	for _, index := range indices {
		password := testHexSeed + "-" + index
		req := &accountv1.CreateRequest{
			Password: password,
		}

		resp, err := accountAPI.Create(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.Account.Id)

		// Check that address is unique
		assert.False(t, addresses[resp.Account.Id], "Duplicate address for index %s: %s", index, resp.Account.Id)
		addresses[resp.Account.Id] = true

		// Check that address starts with 'r'
		assert.Equal(t, uint8('r'), resp.Account.Id[0])
	}

	// Check that all addresses are different
	assert.Equal(t, len(indices), len(addresses))
}
