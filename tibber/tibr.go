package tibber

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/machinebox/graphql"
)

type Client struct {
	APIClient *APIClient
	DBConn    *pgx.Conn
	Logger    *log.Logger
}

func (t *Client) Init() error {
	apiClient, err := NewAPIClient(
		&APIConfig{
			Token: os.Getenv("TIBBER_API_TOKEN"),
			URL:   "https://api.tibber.com/v1-beta/gql",
		},
	)

	if err != nil {
		return fmt.Errorf("error initialising")
	}

	t.APIClient = apiClient

	return nil

}

func (t *Client) Run(ctx context.Context) {
	fmt.Println("Running")

	fmt.Println("Querying")
}

func (t *Client) Query(ctx context.Context) {
	// make a request
	req := graphql.NewRequest(`
		query ($id: ID!, $resolution: EnergyResolution!, $last: Int!) {
			viewer {
				home(id: $id)  {
				  consumption(resolution: $resolution, last: $last) {
					nodes {
					  from
					  to
					  cost
					  unitPrice
					  unitPriceVAT
					  consumption
					  consumptionUnit
					}
				  }
				}
			}
		}
	`)

	// set any variables
	req.Var("id", os.Getenv("TIBBER_HOUSE_ID"))
	req.Var("resolution", "HOURLY")
	req.Var("last", 24)

	// set header fields
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.APIClient.Config.Token)

	// run it and capture the response
	var homeConsumtion HomeConsumptionResponse
	if err := t.APIClient.GQLClient.Run(ctx, req, &homeConsumtion); err != nil {
		log.Fatal(err)
	}
	fmt.Println(homeConsumtion.Viewer.Home.Consumption.Nodes)

}
