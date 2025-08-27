// Package crypto provides cryptographic utilities for XRPL wallet management.
// It includes functions for key derivation, wallet creation, and address generation
// using BIP-44 hierarchical deterministic wallet standards.
package crypto

import (
	"fmt"

	"github.com/CreatureDev/xrpl-go/model/transactions/types"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
)

// Wallet represents an XRPL wallet with address, public key, and private key.
// It provides methods for wallet validation and string representation.
//
// The Wallet struct encapsulates all the information needed to interact with
// the XRPL network, including signing transactions and verifying ownership.
type Wallet struct {
	// Address is the XRPL account address in classic format (starts with 'r').
	// This is the public identifier for the wallet on the XRPL network.
	Address types.Address

	// PublicKey is the hexadecimal representation of the wallet's public key.
	// This is used for transaction validation and verification.
	PublicKey string

	// PrivateKey is the XRPL secret used for signing transactions.
	// This should be kept secure and never exposed in logs or error messages.
	PrivateKey string
}

// NewWallet creates and returns a new Wallet instance.
// It validates the wallet data before returning the instance.
//
// Parameters:
// - address: The XRPL account address
// - publicKey: The hexadecimal public key
// - privateKey: The XRPL secret
//
// Returns a validated Wallet instance or an error if validation fails.
// The function ensures all required fields are present and valid.
func NewWallet(address types.Address, publicKey, privateKey string) (*Wallet, error) {
	w := &Wallet{Address: address, PublicKey: publicKey, PrivateKey: privateKey}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	return w, nil
}

// String returns the string representation of the wallet.
// This is the wallet's address, which is the primary identifier.
//
// Returns the wallet address as a string.
func (w Wallet) String() string {
	return string(w.Address)
}

// Validate checks that the wallet contains valid data.
// It ensures all required fields are present and the address format is correct.
//
// Returns an error if validation fails, or nil if the wallet is valid.
// Validation includes checking for empty fields and XRPL address format.
func (w Wallet) Validate() error {
	if w.Address == "" {
		return fmt.Errorf("wallet address cannot be empty")
	}
	if w.PublicKey == "" {
		return fmt.Errorf("wallet public key cannot be empty")
	}
	if w.PrivateKey == "" {
		return fmt.Errorf("wallet private key cannot be empty")
	}
	return w.Address.Validate()
}

// NewWalletFromExtendedKey creates a new Wallet from an extended key.
// It derives the wallet components using the XRPL-specific key derivation process.
//
// This function is useful when you have an extended key from BIP-44 derivation
// and want to create a complete XRPL wallet.
//
// Parameters:
// - key: An extended key derived from a BIP-44 path
//
// Returns a new Wallet instance or an error if creation fails.
// The function handles the conversion from extended key to wallet components.
func NewWalletFromExtendedKey(key *hdkeychain.ExtendedKey) (*Wallet, error) {
	if key == nil {
		return nil, fmt.Errorf("extended key cannot be nil")
	}
	address, public, private, err := GetXRPLWallet(key)
	if err != nil {
		return nil, err
	}
	return NewWallet(types.Address(address), public, private)
}

// NewWalletFromHexSeed creates a new Wallet from a hexadecimal seed and derivation path.
// It combines the seed derivation and wallet creation into a single function.
//
// This is the most convenient way to create wallets from seed phrases.
// The function handles the complete process from seed to usable wallet.
//
// Parameters:
// - hexSeed: A 64-character hexadecimal string representing the master seed
// - path: The BIP-44 derivation path (e.g., "m/44'/144'/0'/0/0")
//
// Returns a new Wallet instance or an error if creation fails.
// The path should follow BIP-44 standard with XRPL coin type 144.
func NewWalletFromHexSeed(hexSeed string, path string) (*Wallet, error) {
	key, err := GetExtendedKeyFromHexSeedWithPath(hexSeed, path)
	if err != nil {
		return nil, err
	}
	return NewWalletFromExtendedKey(key)
}
