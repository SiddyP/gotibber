package tibber

import (
	"github.com/machinebox/graphql"
)

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
