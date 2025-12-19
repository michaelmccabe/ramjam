package cmd

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var (
	getTimeout int
)

var getCmd = &cobra.Command{
	Use:   "get [url]",
	Short: "Send a GET request to the specified URL",
	Long: `Send a GET request to the specified URL and display the response.
	
Example:
  ramjam get https://api.example.com/users
  ramjam get https://api.example.com/users --timeout 30`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		verbose, _ := cmd.Flags().GetBool("verbose")

		if verbose {
			fmt.Printf("Sending GET request to: %s\n", url)
			fmt.Printf("Timeout: %d seconds\n", getTimeout)
		}

		client := &http.Client{
			Timeout: time.Duration(getTimeout) * time.Second,
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set default user agent
		req.Header.Set("User-Agent", "ramjam-cli")

		if verbose {
			fmt.Println("Sending request...")
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Status Code: %d\n", resp.StatusCode)

		if verbose {
			fmt.Println("\nResponse Headers:")
			for key, values := range resp.Header {
				for _, value := range values {
					fmt.Printf("  %s: %s\n", key, value)
				}
			}
		}

		fmt.Println("\nResponse Body:")
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		fmt.Println(string(body))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().IntVarP(&getTimeout, "timeout", "t", 30, "Request timeout in seconds")
}
