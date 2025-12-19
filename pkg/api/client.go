package api

import (
	"fmt"
	"net/http"
	"time"
)

// Client represents an HTTP API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Verbose    bool
}

// NewClient creates a new API client
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Verbose: false,
	}
}

// SetVerbose enables or disables verbose output
func (c *Client) SetVerbose(verbose bool) {
	c.Verbose = verbose
}

// Get performs a GET request to the specified path
func (c *Client) Get(path string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	
	if c.Verbose {
		fmt.Printf("GET %s\n", url)
	}

	return c.HTTPClient.Get(url)
}
