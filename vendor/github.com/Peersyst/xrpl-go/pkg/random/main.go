package random

import (
	"crypto/rand"
	"io"
)

// Randomizer implements the io.Reader interface. A Randomizer can be used to generate random bytes.
type Randomizer struct {
	io.Reader
}

// NewRandomizer returns a new Randomizer instance. It uses the crypto/rand package to generate random bytes.
func NewRandomizer() Randomizer {
	return Randomizer{
		Reader: rand.Reader,
	}
}

// GenerateBytes generates a n bytes slice of random bytes. It can return an error if the random bytes cannot be read.
// For further information, see the io.Reader interface.
func (r Randomizer) GenerateBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := r.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
