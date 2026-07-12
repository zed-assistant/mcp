package jwt

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	golangjwt "github.com/golang-jwt/jwt/v5"
)

type JWKS struct {
	keys map[string]*rsa.PublicKey
}

type jwksJSON struct {
	Keys []jwksKeyJSON `json:"keys"`
}

type jwksKeyJSON struct {
	Kty string `json:"kty" validate:"required"`
	Kid string `json:"kid" validate:"required"`
	N   string `json:"n" validate:"required"`
	E   string `json:"e" validate:"required"`
}

func ParseJWKS(data []byte) (JWKS, error) {
	var set jwksJSON
	if err := json.Unmarshal(data, &set); err != nil {
		return JWKS{}, fmt.Errorf("invalid JWKS JSON: %w", err)
	}
	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := parseRSAPublicKey(k)
		if err != nil {
			return JWKS{}, fmt.Errorf("invalid RSA key (kid=%q): %w", k.Kid, err)
		}
		keys[k.Kid] = pub
	}
	if len(keys) == 0 {
		return JWKS{}, errors.New("JWKS contains no RSA keys")
	}
	return JWKS{keys: keys}, nil
}

func parseRSAPublicKey(k jwksKeyJSON) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("invalid modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("invalid exponent: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())
	return &rsa.PublicKey{N: n, E: e}, nil
}

type VerifyingOptions struct {
	Secret           []byte
	JWKS             *JWKS
	ExpectedIssuer   string
	ExpectedAudience string
}

func symmetricKeyFunc(options VerifyingOptions) golangjwt.Keyfunc {
	return func(t *golangjwt.Token) (any, error) {
		if _, ok := t.Method.(*golangjwt.SigningMethodHMAC); !ok {
			return nil, golangjwt.ErrSignatureInvalid
		}
		return options.Secret, nil
	}
}

func asymmetricKeyFunc(jwks *JWKS) golangjwt.Keyfunc {
	return func(t *golangjwt.Token) (any, error) {
		if _, ok := t.Method.(*golangjwt.SigningMethodRSA); !ok {
			return nil, golangjwt.ErrSignatureInvalid
		}
		kid, _ := t.Header["kid"].(string)
		if key, ok := jwks.keys[kid]; ok {
			return key, nil
		}
		return nil, fmt.Errorf("no key found for kid %q", kid)
	}
}

func Verify(token string, options VerifyingOptions) (*Claims, error) {
	hasSecret := len(options.Secret) > 0
	hasJWKS := options.JWKS != nil

	if hasSecret == hasJWKS {
		return nil, errors.New("exactly one of Secret (HS256) or JWKS (RS256) must be set")
	}

	var validMethods []string
	var keyFunc golangjwt.Keyfunc
	if hasSecret {
		validMethods = []string{golangjwt.SigningMethodHS256.Alg()}
		keyFunc = symmetricKeyFunc(options)
	} else {
		validMethods = []string{golangjwt.SigningMethodRS256.Alg()}
		keyFunc = asymmetricKeyFunc(options.JWKS)
	}

	parserOptions := []golangjwt.ParserOption{
		golangjwt.WithValidMethods(validMethods),
		golangjwt.WithExpirationRequired(),
	}
	if options.ExpectedIssuer != "" {
		parserOptions = append(parserOptions, golangjwt.WithIssuer(options.ExpectedIssuer))
	}
	if options.ExpectedAudience != "" {
		parserOptions = append(parserOptions, golangjwt.WithAudience(options.ExpectedAudience))
	}

	jwtToken, err := golangjwt.Parse(token, keyFunc, parserOptions...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse JWT token: %w", err)
	}
	if !jwtToken.Valid {
		return nil, errors.New("invalid JWT token")
	}

	mapClaims, ok := jwtToken.Claims.(golangjwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid JWT claims")
	}

	claims := &Claims{}

	if iss, ok := mapClaims["iss"].(string); ok {
		claims.Issuer = iss
	}
	if sub, ok := mapClaims["sub"].(string); ok {
		claims.Subject = sub
	}
	if aud, ok := mapClaims["aud"].(string); ok {
		claims.Audience = []string{aud}
	} else if audList, ok := mapClaims["aud"].([]any); ok {
		for _, audItem := range audList {
			if audStr, ok := audItem.(string); ok {
				claims.Audience = append(claims.Audience, audStr)
			}
		}
	}
	if exp, ok := mapClaims["exp"].(float64); ok {
		claims.Expiry = time.Unix(int64(exp), 0)
	}
	if iat, ok := mapClaims["iat"].(float64); ok {
		claims.IssuedAt = time.Unix(int64(iat), 0)
	}
	if nbf, ok := mapClaims["nbf"].(float64); ok {
		claims.NotBefore = time.Unix(int64(nbf), 0)
	}
	if jti, ok := mapClaims["jti"].(string); ok {
		claims.Jti = jti
	}

	claims.Additional = make(map[string]any)
	for k, v := range mapClaims {
		if k != "iss" && k != "sub" && k != "aud" && k != "exp" && k != "iat" && k != "nbf" && k != "jti" {
			claims.Additional[k] = v
		}
	}

	return claims, nil
}
