package tibber

import (
	"os"

	"github.com/machinebox/graphql"
)

type Query interface {
	Request() interface{}
}

type Consumption struct {
	Request *graphql.Request
}

func (t *Client) Consumption(id string, resolution, string, last int) {
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

}
