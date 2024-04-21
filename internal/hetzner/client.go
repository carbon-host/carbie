package hetzner

import (
	"fmt"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"os"
)

var Client *hcloud.Client

func NewClient() (*hcloud.Client, error) {
	apiKey := os.Getenv("HETZNER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("HETZNER_API_KEY environment variable is not set")
	}

	Client = hcloud.NewClient(hcloud.WithToken(apiKey))
	return Client, nil
}

func GetClient() *hcloud.Client {
	return Client
}
