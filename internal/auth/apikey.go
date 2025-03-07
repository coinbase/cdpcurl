package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type APIKey struct {
	Name       string
	PrivateKey string
}

// UnmarshalJSON allows APIKey to support both the old and new JSON field names.
func (a *APIKey) UnmarshalJSON(data []byte) error {
	// Define a temporary structure with both possible field names.
	var aux struct {
		Name       string `json:"name"`
		ID         string `json:"id"`
		PrivateKey string `json:"privateKey"`
		Secret     string `json:"secret"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Use the old or new field for the API key name.
	if aux.Name != "" {
		a.Name = aux.Name
	} else {
		a.Name = aux.ID
	}

	// Use the old or new field for the API key secret.
	if aux.PrivateKey != "" {
		a.PrivateKey = aux.PrivateKey
	} else {
		a.PrivateKey = aux.Secret
	}

	return nil
}

type apiKeyLoaderConfig struct {
	filename        string
	path            string
	envvars         map[string]string
	envOnly         bool
	fileOnly        bool
	directID        string
	directSecret    string
	useDirectValues bool
}

const (
	nameEnvVar       = "COINBASE_CLOUD_API_KEY_NAME"
	privateKeyEnvVar = "COINBASE_CLOUD_API_PRIVATE_KEY"
	defaultFilename  = ".coinbase_cloud_api_key.json"
)

type LoadAPIKeyOption func(*apiKeyLoaderConfig)

func WithPath(path string) LoadAPIKeyOption {
	return func(c *apiKeyLoaderConfig) {
		c.path = path
	}
}

func WithFileName(filename string) LoadAPIKeyOption {
	return func(c *apiKeyLoaderConfig) {
		c.filename = filename
	}
}

func WithENVVariableNames(name, privateKey string) LoadAPIKeyOption {
	return func(c *apiKeyLoaderConfig) {
		c.envvars = map[string]string{
			nameEnvVar:       name,
			privateKeyEnvVar: privateKey,
		}
	}
}

func WithENVOnly() LoadAPIKeyOption {
	return func(c *apiKeyLoaderConfig) {
		c.fileOnly = false
		c.envOnly = true
	}
}

func WithFileOnly() LoadAPIKeyOption {
	return func(c *apiKeyLoaderConfig) {
		c.fileOnly = true
		c.envOnly = false
	}
}

func WithDirectIDAndSecret(id, secret string) LoadAPIKeyOption {
	return func(c *apiKeyLoaderConfig) {
		c.directID = id
		c.directSecret = secret
		c.useDirectValues = true
	}
}

func LoadAPIKey(options ...LoadAPIKeyOption) (*APIKey, error) {
	c := &apiKeyLoaderConfig{
		filename: defaultFilename,
		envvars: map[string]string{
			nameEnvVar:       nameEnvVar,
			privateKeyEnvVar: privateKeyEnvVar,
		},
	}
	for _, o := range options {
		o(c)
	}

	apiKey := &APIKey{}

	if !c.envOnly {
		if err := c.loadApiKeyFromFile(apiKey); err != nil {
			return nil, fmt.Errorf("api key loader: %w", err)
		}
	}

	if !c.fileOnly {
		c.loadApiKeyFromEnv(apiKey)
	}

	if apiKey.Name == "" || apiKey.PrivateKey == "" {
		return nil, fmt.Errorf("api key loader: could not load api key")
	}

	return apiKey, nil
}

func (c *apiKeyLoaderConfig) loadApiKeyFromFile(a *APIKey) error {
	if c.path != "" {
		f, err := os.Open(c.path)
		if err != nil {
			return fmt.Errorf("file load: %w", err)
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		if err := dec.Decode(a); err != nil {
			return fmt.Errorf("file load: %w", err)
		}

		return nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("file load: %w", err)
	}
	for wd != "" && wd != "/" {
		keyFilepath := path.Join(wd, c.filename)
		f, err := os.Open(keyFilepath)
		if err != nil {
			// Skip if file not accessible.
			wd = path.Dir(wd)
			continue
		}
		defer f.Close()

		dec := json.NewDecoder(f)
		if err := dec.Decode(a); err != nil {
			return fmt.Errorf("file load: %w", err)
		}

		return nil
	}

	return nil
}

func (c *apiKeyLoaderConfig) loadApiKeyFromEnv(a *APIKey) {
	if c.useDirectValues {
		if a.Name == "" {
			a.Name = c.directID
		}
		if a.PrivateKey == "" {
			a.PrivateKey = c.directSecret
		}
		return
	}

	if a.Name == "" {
		a.Name = os.Getenv(c.envvars[nameEnvVar])
	}
	if a.PrivateKey == "" {
		a.PrivateKey = os.Getenv(c.envvars[privateKeyEnvVar])
	}
}
