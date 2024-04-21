package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func ListServers(client *hcloud.Client) ([]*hcloud.Server, error) {
	ctx := context.Background()

	opts := hcloud.ServerListOpts{}
	servers, err := client.Server.AllWithOpts(ctx, opts)
	if err != nil {
		return nil, err
	}

	return servers, nil
}
