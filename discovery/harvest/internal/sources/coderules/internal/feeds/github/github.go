package github

import (
	"context"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
)

type Content = feeds.GHContent

type Client struct {
	feeds *feeds.Client
}

func NewClient(fc *feeds.Client) *Client {
	return &Client{feeds: fc}
}

func NewClientHTTP() *Client {
	return &Client{feeds: feeds.NewClient("", nil)}
}

func (c *Client) ListDir(ctx context.Context, owner, repo, path string) ([]Content, error) {
	return feeds.GitHubListDir(ctx, c.feeds, owner, repo, path)
}

func (c *Client) FetchText(ctx context.Context, owner, repo, ref, path string) ([]byte, error) {
	return feeds.GitHubFetchRaw(ctx, c.feeds, owner, repo, ref, path)
}
