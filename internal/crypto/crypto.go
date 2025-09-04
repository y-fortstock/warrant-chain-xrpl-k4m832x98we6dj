// Package crypto provides cryptographic utilities for XRPL wallet management.
// It includes functions for key derivation, wallet creation, and address generation
// using BIP-44 hierarchical deterministic wallet standards.
package crypto

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	ac "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/keypairs"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
)

// GetExtendedKeyFromHexSeedWithPath creates an extended key from a hexadecimal seed string
// and derives it along the specified BIP-44 derivation path.
//
// This function is the main entry point for creating XRPL wallets from seed phrases.
// It decodes the hex seed and calls GetExtendedKeyFromSeedWithPath for the actual derivation.
//
// Parameters:
// - hexSeed: A 64-character hexadecimal string representing the master seed
// - path: The BIP-44 derivation path (e.g., "m/44'/144'/0'/0/0")
//
// Returns an extended key that can be used to derive child keys, or an error if derivation fails.
// The path should follow BIP-44 standard with XRPL coin type 144.
func GetExtendedKeyFromHexSeedWithPath(hexSeed string, path string) (*hdkeychain.ExtendedKey, error) {
	seed, err := hex.DecodeString(hexSeed)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex seed: %w", err)
	}
	return GetExtendedKeyFromSeedWithPath(seed, path)
}

// GetExtendedKeyFromSeedWithPath creates an extended key from raw seed bytes
// and derives it along the specified BIP-44 derivation path.
//
// This function implements the core BIP-44 derivation logic for XRPL wallets.
// It creates a master key and then derives child keys according to the specified path.
//
// Parameters:
// - seed: Raw bytes representing the master seed
// - path: The BIP-44 derivation path (e.g., "m/44'/144'/0'/0/0")
//
// Returns an extended key derived along the specified path, or an error if derivation fails.
// The function uses MainNet parameters for key derivation.
func GetExtendedKeyFromSeedWithPath(seed []byte, path string) (*hdkeychain.ExtendedKey, error) {
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create master key: %w", err)
	}

	derivationPath, err := parseDerivationPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse derivation path: %w", err)
	}

	currentKey := masterKey
	for i, childIndex := range derivationPath {
		currentKey, err = currentKey.Derive(childIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to derive key at level %d (index %d): %w", i, childIndex, err)
		}
	}

	return currentKey, nil
}

// parseDerivationPath parses a BIP-44 derivation path string into an array of indices.
// It handles both hardened and normal derivation components.
//
// The function supports paths with or without the "m/" prefix and handles
// hardened derivation (indicated by apostrophes) according to BIP-44 standards.
//
// Parameters:
// - path: A BIP-44 derivation path string (e.g., "m/44'/144'/0'/0/0")
//
// Returns an array of uint32 indices representing the derivation path, or an error if parsing fails.
// Hardened derivation indices are offset by HardenedKeyStart (0x80000000).
func parseDerivationPath(path string) ([]uint32, error) {
	if path == "" {
		return nil, fmt.Errorf("path is empty")
	}

	if len(path) >= 2 && path[:2] == "m/" {
		path = path[2:]
	}

	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid path format")
	}

	derivationPath := make([]uint32, len(parts))
	for i, part := range parts {
		hardened := false
		if strings.HasSuffix(part, "'") {
			hardened = true
			part = part[:len(part)-1]
		}

		index, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid path component %s: %w", part, err)
		}

		if hardened {
			derivationPath[i] = hdkeychain.HardenedKeyStart + uint32(index)
		} else {
			derivationPath[i] = uint32(index)
		}
	}

	return derivationPath, nil
}

// GetXRPLWallet creates a complete XRPL wallet from an extended key.
// It generates the address, public key, and private key needed for XRPL transactions.
//
// This function is the main interface for creating XRPL wallets from derived keys.
// It handles the conversion from Bitcoin-style keys to XRPL-specific formats.
//
// Parameters:
// - key: An extended key derived from a BIP-44 path
//
// Returns the wallet address, public key (hex), private key (secret), and any error that occurred.
// The address is in XRPL classic address format (starts with 'r').
func GetXRPLWallet(key *hdkeychain.ExtendedKey) (address string, public string, private string, err error) {
	secret, err := getXRPLSecret(key)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get secret from key: %w", err)
	}

	privKey, pubKeyHex, err := keypairs.DeriveKeypair(secret, false)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to derive keypair: %w", err)
	}

	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to decode public key: %w", err)
	}

	accountID := ac.Sha256RipeMD160(pubKeyBytes)
	address, err = ac.Encode(accountID, []byte{ac.AccountAddressPrefix}, ac.AccountAddressLength)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to encode account ID: %w", err)
	}
	return address, pubKeyHex, privKey, nil
}

// getXRPLSecret converts a Bitcoin extended key to XRPL secret format.
// This function handles the conversion from Bitcoin private key format to XRPL secret encoding.
//
// XRPL uses a specific encoding format for private keys that includes version bytes
// and checksums for validation and error detection.
//
// Parameters:
// - key: An extended key containing the private key information
//
// Returns the XRPL secret string that can be used for keypair derivation, or an error if conversion fails.
// The secret is encoded with XRPL-specific version bytes and checksums.
func getXRPLSecret(key *hdkeychain.ExtendedKey) (string, error) {
	privKey, err := key.ECPrivKey()
	if err != nil {
		return "", fmt.Errorf("failed to get private key: %w", err)
	}

	privKeyBytes := privKey.Serialize()

	secret, err := ac.Encode(privKeyBytes,
		[]byte{0x01, 0xe1, 0x4b},
		32,
	)
	if err != nil {
		return "", fmt.Errorf("failed to encode secret: %w", err)
	}
	return secret, nil
}

// NewWallet creates a new Wallet from address, public key, and private key
func NewWallet(address types.Address, publicKey string, privateKey string) (*wallet.Wallet, error) {
	if address == "" {
		return nil, fmt.Errorf("system account is not set")
	}
	if publicKey == "" {
		return nil, fmt.Errorf("system public key is not set")
	}
	if privateKey == "" {
		return nil, fmt.Errorf("system secret key is not set")
	}

	return &wallet.Wallet{
		ClassicAddress: address,
		PublicKey:      publicKey,
		PrivateKey:     privateKey,
	}, nil
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
func NewWalletFromExtendedKey(key *hdkeychain.ExtendedKey) (*wallet.Wallet, error) {
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
func NewWalletFromHexSeed(hexSeed string, path string) (*wallet.Wallet, error) {
	key, err := GetExtendedKeyFromHexSeedWithPath(hexSeed, path)
	if err != nil {
		return nil, err
	}
	return NewWalletFromExtendedKey(key)
}
