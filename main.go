package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/coinbase/cdpcurl/internal/auth"
	"github.com/coinbase/cdpcurl/transport"
	"github.com/spf13/cobra"
)

var (
	version = "v0.0.6"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Print the version number of cdpcurl",
	Long:    `All software has versions. This is cdpcurl's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

func main() {
	var data, method, apiKeyPath, header string
	var versionFlag bool
	var id, secret string

	cmd := &cobra.Command{
		Use:  "cdpcurl [flags] [URL]",
		Args: cobra.MinimumNArgs(0), // Allow zero arguments to handle -v
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				fmt.Println(version)
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("URL is required unless using -v")
			}

			opts := []transport.Option{}
			if apiKeyPath != "" {
				opts = append(opts, transport.WithAPIKeyLoaderOption(transport.WithPath(apiKeyPath)))
			}

			// Add options for id and secret if they are provided
			if id != "" && secret != "" {
				opts = append(opts, transport.WithAPIKeyLoaderOption(auth.WithDirectIDAndSecret(id, secret)))
			}

			authTransport, err := transport.New("", http.DefaultTransport, opts...)
			if err != nil {
				return err
			}

			req, err := http.NewRequest(method, args[0], bytes.NewBufferString(data))
			if err != nil {
				return err
			}

			if method == http.MethodPost && header == "" {
				req.Header.Set("Content-Type", "application/json")
			}

			if header != "" {
				var headers map[string]string
				if err := json.Unmarshal([]byte(header), &headers); err != nil {
					return err
				}

				for key, value := range headers {
					req.Header.Set(key, value)
				}
			}

			client := http.Client{Transport: authTransport}

			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			// Print HTTP status code and response body
			fmt.Println(resp.Status)
			fmt.Println(string(body))
			return nil
		},
	}

	cmd.Flags().StringVarP(&data, "data", "d", "", "HTTP Body")
	cmd.Flags().StringVarP(&apiKeyPath, "api-key-path", "k", "", "API Key Path")
	cmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP Method")
	cmd.Flags().StringVarP(&header, "header", "H", "", "HTTP Header")
	cmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Print the version number and exit")
	cmd.Flags().StringVarP(&id, "id", "i", "", "API Key ID (only works with Ed25519 keys)")
	cmd.Flags().StringVarP(&secret, "secret", "s", "", "API Key Secret (only works with Ed25519 keys)")

	cmd.AddCommand(versionCmd)

	// Custom usage template
	customUsage := `Usage:
  {{.UseLine}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
`

	cmd.SetUsageTemplate(customUsage)

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
