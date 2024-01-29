package tibber

import (
	"context"

	"github.com/machinebox/graphql"
)

type APIClienter interface {
	Run(ctx context.Context, req *graphql.Request, resp interface{}) error
}

type APIConfig struct {
	Token string
	URL   string
}

type APIClient struct {
	Config    *APIConfig
	GQLClient *graphql.Client
}

func NewAPIClient(config *APIConfig) *APIClient {
	// ensure token is set in config
	if config.Token == "" {
		panic("TIBBER_API_TOKEN not set")
	}

	if config.URL == "" {
		panic("TIBBER_API_URL not set")
	}

	client := &APIClient{
		Config:    config,
		GQLClient: graphql.NewClient(config.URL),
	}
	return client
}

func (c *APIClient) Run(ctx context.Context, req *graphql.Request, resp interface{}) error {
	req.Header.Set("Authorization", "Bearer "+c.Config.Token)
	return c.GQLClient.Run(ctx, req, resp)
}
