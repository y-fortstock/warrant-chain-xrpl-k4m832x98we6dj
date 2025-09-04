package transaction

var _ Tx = (*FlatTransaction)(nil)

type FlatTransaction map[string]interface{}

func (f FlatTransaction) TxType() TxType {
	txType, ok := f["TransactionType"].(string)
	if !ok {
		return TxType("")
	}
	return TxType(txType)
}
