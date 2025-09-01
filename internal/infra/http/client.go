package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
	"github.com/subconverter/subconverter-go/internal/pkg/errors"
)

// Client represents an HTTP client
type Client struct {
	client *req.Client
}

// NewClient creates a new HTTP client
func NewClient() *Client {
	return &Client{
		client: req.C().
			SetTimeout(30 * time.Second).
			SetUserAgent("SubConverter-Go/1.0").
			EnableInsecureSkipVerify(),
	}
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	resp, err := c.client.R().
		SetContext(ctx).
		Get(url)
	
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch URL")
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, errors.BadRequest("FETCH_FAILED", fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status))
	}
	
	return resp.Bytes(), nil
}

// Health checks the client health
func (c *Client) Health(ctx context.Context) error {
	_, err := c.client.R().
		SetContext(ctx).
		Head("http://httpbin.org/headers")
	
	return err
}

// SetTimeout sets the timeout for requests
func (c *Client) SetTimeout(timeout time.Duration) {
	c.client.SetTimeout(timeout)
}