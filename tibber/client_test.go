package tibber

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientToken(t *testing.T) {
	assert := assert.New(t)

	client := APIClient{
		Config: &APIConfig{
			Token: "dummy_token",
		},
	}
	assert.Equal(client.Config.Token, "dummy_token", "The two tokens should be the same.")
}

// func TestSplit(t *testing.T) {
//     tests := map[string]struct {
//         input string
//         sep   string
//         want  []string
//     }{
//         "simple":       {input: "a/b/c", sep: "/", want: []string{"a", "b", "c"}},
//         "wrong sep":    {input: "a/b/c", sep: ",", want: []string{"a/b/c"}},
//         "no sep":       {input: "abc", sep: "/", want: []string{"abc"}},
//         "trailing sep": {input: "a/b/c/", sep: "/", want: []string{"a", "b", "c"}},
//     }

//     for name, tc := range tests {
//         got := Split(tc.input, tc.sep)
//         if !reflect.DeepEqual(tc.want, got) {
//             t.Fatalf("%s: expected: %v, got: %v", name, tc.want, got)
//         }
//     }
// }
