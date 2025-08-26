package types

import "fmt"

type Hash192 string

func (h Hash192) Validate() error {
	if h == "" {
		return fmt.Errorf("hash192 value not set")
	}
	if len(h) != 48 {
		return fmt.Errorf("hash192 length was not expected 48 characters")
	}
	return nil
}
