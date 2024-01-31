package reporter

import (
	"context"
	"net/url"

	"golang.org/x/xerrors"

	"github.com/coder/coder/v2/agent/proto"
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
	client, err := agent.RPC(ctx)
	if err != nil {
		return agentsdk.Manifest{}, xerrors.Errorf("rpc: %w", err)
	}

	pm, err := client.GetManifest(ctx, &proto.GetManifestRequest{})
	if err != nil {
		return agentsdk.Manifest{}, xerrors.Errorf("get manifest: %w", err)
	}
	return agentsdk.ManifestFromProto(pm)
}

func (c *client) PostJFrogXrayScan(ctx context.Context, req codersdk.JFrogXrayScan) error {
	return c.coder.PostJFrogXrayScan(ctx, req)
}
