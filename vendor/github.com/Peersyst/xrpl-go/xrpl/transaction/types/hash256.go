package types

type Hash256 string

func (h *Hash256) String() string {
	return string(*h)
}
