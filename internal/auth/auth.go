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

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
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
	var (
		key any
		alg jose.SignatureAlgorithm
	)

	keyStr := a.apiKey.PrivateKey

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
	
	// Prepare the JOSE signer options.
	jsonSignatureOptions := &jose.SignerOptions{}
	// Set the "typ" header to "JWT"
	jsonSignatureOptions.WithType("JWT")
	// Set the "kid" header to the API key name
	jsonSignatureOptions.WithHeader("kid", a.apiKey.Name)
	// Set a custom nonce source to generate a unique nonce for each token
	jsonSignatureOptions.NonceSource = nonceSource{}

	// Create the JOSE signer with the appropriate algorithm.
	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: alg, Key: key},
		jsonSignatureOptions,
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
	jwtString, err := jwt.Signed(sig).Claims(claims).Serialize()
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
