package random

import "crypto/rand"

type Random struct{}

func NewRandom() *Random {
	return &Random{}
}

func (s *Random) RandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
