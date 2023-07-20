package multifaucet

import (
	"context"
	"net/url"
	"strings"

	"github.com/alt-research/operator-kit/must"
	"github.com/carlmjohnson/requests"
)

type Client struct {
	url string
	key string
}

func NewClient(url, key string) *Client {
	return &Client{url, key}
}

func (c *Client) builder(path string) *requests.Builder {
	return requests.URL(must.Two(url.JoinPath(c.url, path))).Param("key", c.key)
}

func (c *Client) GetAll(ctx context.Context) ([]Network, error) {
	data := map[string][]Network{"networks": {}}
	err := c.builder("api/networks").Method("GET").ToJSON(&data).Fetch(ctx)
	return data["networks"], err
}

func (c *Client) DeleteAll(ctx context.Context) ([]Network, error) {
	data := map[string][]Network{"networks": {}}
	rst := Result{}
	err := c.builder("api/networks").Method("DELETE").ToJSON(&data).ErrorJSON(&rst).Fetch(ctx)
	if rst.Err != "" {
		return data["networks"], &rst
	}
	return data["networks"], err
}

func (c *Client) Delete(ctx context.Context, chainID string) error {
	rst := Result{}
	err := c.builder("api/network/" + chainID).Method("DELETE").ToJSON(&rst).Fetch(ctx)
	if rst.Err != "" {
		return &rst
	}
	return err
}

func (c *Client) Get(ctx context.Context, chainID string) (*Network, error) {
	data := &Network{}
	rst := Result{}
	err := c.builder("api/network/" + chainID).Method("GET").ToJSON(data).ErrorJSON(&rst).Fetch(ctx)
	if rst.Err != "" {
		return data, &rst
	}
	if err != nil && strings.Contains(err.Error(), "unexpected end of JSON input") {
		return nil, ErrNotFound
	}
	return data, err
}

func (c *Client) Upsert(ctx context.Context, network *Network) error {
	rst := Result{}
	if network.ERC20Tokens == nil {
		network.ERC20Tokens = []string{}
	}
	if len(network.ERC20Tokens) == 0 && !network.SupportNative {
		network.SupportNative = true
	}
	err := c.builder("api/network").Method("POST").
		BodyJSON(map[string]*Network{"network": network}).
		ToJSON(&rst).
		Fetch(ctx)
	if rst.Err != "" {
		return &rst
	}
	return err
}
