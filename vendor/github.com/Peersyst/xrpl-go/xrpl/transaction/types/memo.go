package types

type MemoWrapper struct {
	Memo Memo
}

type Memo struct {
	MemoData   string `json:",omitempty"`
	MemoFormat string `json:",omitempty"`
	MemoType   string `json:",omitempty"`
}

func (mw *MemoWrapper) Flatten() map[string]interface{} {
	if mw.Memo != (Memo{}) {
		flattened := make(map[string]interface{})
		flattened["Memo"] = mw.Memo.Flatten()
		return flattened
	}
	return nil
}

func (m *Memo) Flatten() map[string]interface{} {
	flattened := make(map[string]interface{})

	if m.MemoData != "" {
		flattened["MemoData"] = m.MemoData
	}

	if m.MemoFormat != "" {
		flattened["MemoFormat"] = m.MemoFormat
	}

	if m.MemoType != "" {
		flattened["MemoType"] = m.MemoType
	}

	return flattened
}
