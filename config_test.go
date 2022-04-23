package ingester

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LoadConfig(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{
			name: "example",
			env: `[
				{
					"name": "paystack",
					"verifier_config": {
						"type": "hmac",
						"header": "X-Paystack-Signature",
						"hash": "SHA512",
						"secret": "Paystack Secret"
					}

				}
			]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			t.Setenv(CONFIG_ENV, tc.env)

			err := LoadConfig(CONFIG_ENV)
			require.NoError(t, err)
		})
	}
}
