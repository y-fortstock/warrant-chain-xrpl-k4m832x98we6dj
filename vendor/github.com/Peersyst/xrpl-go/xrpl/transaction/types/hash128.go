package types

type Hash128 string

func (h *Hash128) String() string {
	return string(*h)
}
