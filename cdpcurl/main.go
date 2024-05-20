package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"encoding/json"
	"github.com/coinbase/cdpcurl/transport"
	"github.com/spf13/cobra"
)

var (
	version = "v0.0.1"
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
	cmd := &cobra.Command{
		Use:  "cdpcurl [flags] [URL]",
		Args: cobra.MinimumNArgs(1), // Ensure at least one argument is provided
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := []transport.Option{}
			if apiKeyPath != "" {
				opts = append(opts, transport.WithAPIKeyLoaderOption(transport.WithPath(apiKeyPath)))
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

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			fmt.Println(string(body))
			return nil
		},
	}

	cmd.Flags().StringVarP(&data, "data", "d", "", "HTTP Body")
	cmd.Flags().StringVarP(&apiKeyPath, "api-key-path", "k", "", "API Key Path")
	cmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP Method")
	cmd.Flags().StringVarP(&header, "header", "H", "", "HTTP Header")
	cmd.PersistentFlags().BoolP("version", "v", false, "Print the version number and exit")

	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Println(version)
			os.Exit(0)
		}
	}

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
