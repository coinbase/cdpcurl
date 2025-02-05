package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

type APIKeyClaims struct {
	*jwt.Claims
	URIs []string `json:"uris"`
}

type Authenticator struct {
	apiKey APIKey
}

func New(apiKeyOpts ...LoadAPIKeyOption) (*Authenticator, error) {
	apiKey, err := LoadAPIKey(apiKeyOpts...)
	if err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}
	return &Authenticator{
		apiKey: *apiKey,
	}, nil
}

func NewFromConfig(apiKey APIKey) *Authenticator {
	return &Authenticator{
		apiKey: apiKey,
	}
}

func (a *Authenticator) BuildJWT(service string, uris []string) (string, error) {
	keyStr := a.apiKey.PrivateKey
	var (
		key interface{}
		alg jose.SignatureAlgorithm
	)

	// If the key starts with a PEM header, parse it as an ECDSA key.
	if strings.HasPrefix(keyStr, "-----BEGIN") {
		block, _ := pem.Decode([]byte(keyStr))
		if block == nil {
			return "", fmt.Errorf("jwt: could not decode PEM private key")
		}
		ecdsaKey, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("jwt: error parsing EC private key: %w", err)
		}
		key = ecdsaKey
		alg = jose.ES256
	} else {
		// Otherwise, assume the key is a Base64 encoded Ed25519 private key.
		decodedKey, err := base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return "", fmt.Errorf("jwt: error base64 decoding key: %w", err)
		}
		if len(decodedKey) != ed25519.PrivateKeySize {
			return "", fmt.Errorf("jwt: invalid Ed25519 private key length: got %d, expected %d", len(decodedKey), ed25519.PrivateKeySize)
		}
		key = ed25519.PrivateKey(decodedKey)
		alg = jose.EdDSA
	}

	// Create the JOSE signer with the appropriate algorithm.
	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: alg, Key: key},
		(&jose.SignerOptions{NonceSource: nonceSource{}}).
			WithType("JWT").
			WithHeader("kid", a.apiKey.Name),
	)
	if err != nil {
		return "", fmt.Errorf("jwt: error creating signer: %w", err)
	}

	// Build the JWT claims.
	claims := &APIKeyClaims{
		Claims: &jwt.Claims{
			Subject:   a.apiKey.Name,
			Issuer:    "coinbase-cloud",
			NotBefore: jwt.NewNumericDate(time.Now()),
			Expiry:    jwt.NewNumericDate(time.Now().Add(1 * time.Minute)),
			Audience:  jwt.Audience{service},
		},
		URIs: uris,
	}

	// Serialize the JWT.
	jwtString, err := jwt.Signed(sig).Claims(claims).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("jwt: error serializing token: %w", err)
	}
	return jwtString, nil
}

var max = big.NewInt(math.MaxInt64)

type nonceSource struct{}

func (n nonceSource) Nonce() (string, error) {
	r, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return r.String(), nil
}
