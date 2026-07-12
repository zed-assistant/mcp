package jwt

import (
	"maps"

	golangjwt "github.com/golang-jwt/jwt/v5"
)

type SigningOptions struct {
	Secret []byte
	KeyID  string
}

func Sign(claims Claims, options SigningOptions) (string, error) {
	cl := golangjwt.MapClaims{}
	if claims.Issuer != "" {
		cl["iss"] = claims.Issuer
	}
	if claims.Subject != "" {
		cl["sub"] = claims.Subject
	}
	if len(claims.Audience) > 0 {
		cl["aud"] = claims.Audience
	}
	if !claims.Expiry.IsZero() {
		cl["exp"] = claims.Expiry.Unix()
	}
	if !claims.IssuedAt.IsZero() {
		cl["iat"] = claims.IssuedAt.Unix()
	}
	if !claims.NotBefore.IsZero() {
		cl["nbf"] = claims.NotBefore.Unix()
	}
	if claims.Jti != "" {
		cl["jti"] = claims.Jti
	}
	maps.Copy(cl, claims.Additional)

	token := golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, cl)
	if options.KeyID != "" {
		token.Header["kid"] = options.KeyID
	}

	return token.SignedString(options.Secret)
}
