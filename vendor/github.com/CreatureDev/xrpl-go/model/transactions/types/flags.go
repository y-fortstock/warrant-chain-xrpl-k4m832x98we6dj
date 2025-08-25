package types

const (
	// Transaction flags
	FtfBurnable            Flag = 0x1
	FtfOnlyXrp                  = 0x2
	FtfTransferable             = 0x8
	FtfNoDirectRipple           = 0x10000
	FtfPartialPayment           = 0x20000
	FtfLimitQuality             = 0x40000
	FtfRequireDestTag           = 0x10000
	FtfOptionalDestTag          = 0x20000
	FtfRequireAuth              = 0x40000
	FtfOptionalAuth             = 0x80000
	FtfDisallowXRP              = 0x100000
	FtfAllowXRP                 = 0x200000
	FtfLPToken                  = 0x10000
	FtfTwoAsset                 = 0x100000
	FtfTwoAssetIfEmpty          = 0x800000
	FtfSingleAsset              = 0x80000
	FtfOneAssetLPToken          = 0x200000
	FtfLimitLPToken             = 0x400000
	FtfWithdrawAll              = 0x20000
	FtfOneAssetWithdrawAll      = 0x40000
	FtfSellNFToken              = 0x1

	// Account Set Flags
	FasfAccountTxnID                 = 5
	FasfAllowTrustLineClawback       = 16
	FasfAuthorizedNFTokenMinter      = 10
	FasfDefaultRipple                = 8
	FasfDepositAuth                  = 9
	FasfDisableMaster                = 4
	FasfDisallowIncomingCheck        = 13
	FasfDisallowIncomingNFTokenOffer = 12
	FasfDisallowIncomingPayChan      = 14
	FasfDisallowIncomingTrustline    = 15
	FasfDisallowXRP                  = 3
	FasfGlobalFreeze                 = 7
	FasfNoFreeze                     = 6
	FasfRequireAuth                  = 2
	FasfRequireDest                  = 1

	// MPToken Authorize Flags

	// TfMPTUnauthorize
	// If set and transaction is submitted by a holder, it indicates that the holder no
	// longer wants to hold the MPToken, which will be deleted as a result. If the the holder's
	// MPToken has non-zero balance while trying to set this flag, the transaction will fail. On
	// the other hand, if set and transaction is submitted by an issuer, it would mean that the
	// issuer wants to unauthorize the holder (only applicable for allow-listing),
	// which would unset the lsfMPTAuthorized flag on the MPToken.
	TfMPTUnauthorize = 0x00000001

	// MPToken IssuanceCreate Flags

	// TfMPTCanLock
	// If set, indicates that the MPT can be locked both individually and globally.
	// If not set, the MPT cannot be locked in any way.
	TfMPTCanLock = 0x00000002
	// TfMPTRequireAuth
	// If set, indicates that individual holders must be authorized.
	// This enables issuers to limit who can hold their assets.
	TfMPTRequireAuth = 0x00000004
	// TfMPTCanEscrow
	// If set, indicates that individual holders can place their balances into an escrow.
	TfMPTCanEscrow = 0x00000008
	// TfMPTCanTrade
	// If set, indicates that individual holders can trade their balances
	// using the XRP Ledger DEX or AMM.
	TfMPTCanTrade = 0x00000010
	// TfMPTCanTransfer
	// If set, indicates that tokens may be transferred to other accounts
	// that are not the issuer.
	TfMPTCanTransfer = 0x00000020
	// TfMPTCanClawback
	// If set, indicates that the issuer may use the Clawback transaction
	// to clawback value from individual holders.
	TfMPTCanClawback = 0x00000040

	// MPToken IssuanceSet Flags

	// TfMPTLock
	// If set, indicates that issuer locks the MPT
	TfMPTLock = 0x00000001
	// TfMPTUnlock
	// If set, indicates that issuer unlocks the MPT
	TfMPTUnlock = 0x00000002
)

// FlagsI is an interface for types that can be converted to a uint.
type FlagsI interface {
	ToUint() uint32
}

type Flag uint32

func (f *Flag) ToUint() uint32 {
	return uint32(*f)
}

// SetFlag is a helper function that allocates a new uint value
// to store v and returns a pointer to it.
func SetFlag(v uint32) *Flag {
	p := new(uint32)
	*p = v
	return (*Flag)(p)
}

func NewFlag() *Flag {
	return SetFlag(0)
}

func (f *Flag) SetFlag(v Flag) *Flag {
	*f = *f | v
	return f
}

func (f *Flag) ClearFlag(v Flag) *Flag {
	*f = *f &^ v
	return f
}

func (f *Flag) ToggleFlag(v Flag) *Flag {
	*f = *f ^ v
	return f
}

func (f *Flag) HasFlag(v Flag) bool {
	if *f&v != 0 {
		return true
	}
	return false
}
