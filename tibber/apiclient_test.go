package tibber

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateClient(t *testing.T) {
	assert := assert.New(t)

	client := NewAPIClient(&APIConfig{
		Token: "dummy_token",
		URL:   "https://api.tibber.com/v1-beta/gql",
	})
	assert.Equal(client.Config.Token, "dummy_token", "The two tokens should be the same.")
	assert.Equal(client.Config.URL, "https://api.tibber.com/v1-beta/gql", "The two URLs should be the same.")
}
