package types

type Address string

func (a Address) String() string {
	return string(a)
}
