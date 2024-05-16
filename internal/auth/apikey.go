package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type APIKey struct {
	Name       string `json:"name"`
	PrivateKey string `json:"privateKey"`
}

type apiKeyLoaderConfig struct {
	filename string
	path     string
	envvars  map[string]string
	envOnly  bool
	fileOnly bool
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
		defer f.Close()
		if err != nil {
			return fmt.Errorf("file load: %w", err)
		}

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
		wd = path.Dir(wd)
		f, err := os.Open(keyFilepath)
		if err != nil {
			// skip if file not accessable
			continue
		}

		dec := json.NewDecoder(f)
		err = dec.Decode(a)
		if err != nil {
			return fmt.Errorf("file load: %w", err)
		}

		return nil
	}

	return nil
}

func (c *apiKeyLoaderConfig) loadApiKeyFromEnv(a *APIKey) {
	if a.Name == "" {
		a.Name = os.Getenv(c.envvars[nameEnvVar])
	}
	if a.PrivateKey == "" {
		a.PrivateKey = os.Getenv(c.envvars[privateKeyEnvVar])
	}
}
