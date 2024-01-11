package tibber

import (
	"fmt"

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

func NewAPIClient(config *APIConfig) (*APIClient, error) {
	client := &APIClient{
		Config:    config,
		GQLClient: graphql.NewClient(config.URL),
	}

	if config.Token == "" {
		return client, fmt.Errorf("missing token")
	}

	return client, nil
}
