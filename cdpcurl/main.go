package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/coinbase/cdpcurl/transport"

	"github.com/spf13/cobra"

	"encoding/json"
)

func main() {
	var data, method, apiKeyPath, header string
	cmd := &cobra.Command{
		Use: "cdpcurl [URL]",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return err
			}

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
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
