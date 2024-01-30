package reporter

import (
	"context"
	"net/url"

	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/codersdk/agentsdk"
)

type CoderClient interface {
	AgentManifest(ctx context.Context, token string) (agentsdk.Manifest, error)
	PostJFrogXrayScan(ctx context.Context, req codersdk.JFrogXrayScan) error
}

func NewClient(coderURL *url.URL, token string) CoderClient {
	c := codersdk.New(coderURL)
	c.SetSessionToken(token)
	return &client{
		coder: c,
	}
}

type client struct {
	coder *codersdk.Client
}

func (c *client) AgentManifest(ctx context.Context, token string) (agentsdk.Manifest, error) {
	agent := agentsdk.New(c.coder.URL)
	agent.SetSessionToken(token)
	return agent.Manifest(ctx)
}

func (c *client) PostJFrogXrayScan(ctx context.Context, req codersdk.JFrogXrayScan) error {
	return c.coder.PostJFrogXrayScan(ctx, req)
}
