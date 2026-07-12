package jwt

import "time"

type Claims struct {
	Issuer     string
	Subject    string
	Audience   []string
	Expiry     time.Time
	IssuedAt   time.Time
	NotBefore  time.Time
	Jti        string
	Additional map[string]any
}
