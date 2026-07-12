package configuration

import (
	"encoding/base64"
	"fmt"
)

type B64EncodedBytes []byte

func (b *B64EncodedBytes) UnmarshalText(text []byte) error {
	decoded, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return fmt.Errorf("failed to decode base64 string: %w", err)
	}
	*b = decoded
	return nil
}
