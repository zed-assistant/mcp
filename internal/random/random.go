package random

import (
	"crypto/rand"
	"encoding/hex"
)

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

func (s *Random) RandomBytesHex(length int) (string, error) {
	bytes, err := s.RandomBytes(length)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
