package ingester

import (
	"encoding/hex"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// initialize environment variables
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "1")
	os.Exit(m.Run())
}

func Test_HmacVerifier_VerifyRequest(t *testing.T) {
	tests := map[string]struct {
		opts          *HmacConfig
		payload       []byte
		requestFn     func(t *testing.T) *http.Request
		expectedError error
	}{
		"invalid_signature": {
			opts: &HmacConfig{
				Header: "X-Convoy-Signature",
				Hash:   "SHA512",
				Secret: "Convoy",
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				hash := hex.EncodeToString([]byte(`Obviously wrong hash`))

				req.Header.Add("X-Convoy-Signature", hash)
				return req
			},
			expectedError: ErrHashDoesNotMatch,
		},
		"invalid_hex": {
			opts: &HmacConfig{
				Header: "X-Convoy-Signature",
				Hash:   "SHA512",
				Secret: "Convoy",
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				hash := "Hash with characters outside hex"

				req.Header.Add("X-Convoy-Signature", hash)
				return req
			},
			expectedError: ErrCannotDecodeMACHeader,
		},
		"empty_signature": {
			opts: &HmacConfig{
				Header: "X-Convoy-Signature",
				Hash:   "SHA512",
				Secret: "Convoy",
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				req.Header.Add("X-Convoy-Signature", "")
				return req
			},
			expectedError: ErrSignatureCannotBeEmpty,
		},
		"valid_request": {
			opts: &HmacConfig{
				Header: "X-Convoy-Signature",
				Hash:   "SHA512",
				Secret: "Convoy",
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				hash := "83306382f5361d35351d6de45998f23b52f40bcf96befe4e92f137c0f1" +
					"bf4a7119388b238d8f9d502ac77e6f1a8849a4778272667ed88d530cac8050bd1fee2d"

				req.Header.Add("X-Convoy-Signature", hash)
				req.Header.Add("X-Forwarded-For", "52.52.52.52")
				return req
			},
			expectedError: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange.
			v := HmacVerifier{tc.opts}
			req := tc.requestFn(t)

			// Assert.
			err := v.VerifyRequest(req, tc.payload)

			// Act.
			require.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func Test_BasicAuthVerifier_VerifyRequest(t *testing.T) {
	tests := map[string]struct {
		opts          *BasicAuthConfig
		payload       []byte
		requestFn     func(t *testing.T, c *BasicAuthConfig) *http.Request
		expectedError error
	}{
		"valid_request": {
			opts: &BasicAuthConfig{
				Username: "convoy-ingester",
				Password: "convoy-ingester",
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T, c *BasicAuthConfig) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				req.SetBasicAuth(c.Username, c.Password)

				return req
			},
			expectedError: nil,
		},
		"invalid_credentials": {
			opts: &BasicAuthConfig{
				Username: "convoy-ingester",
				Password: "convoy-ingester",
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T, c *BasicAuthConfig) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				req.SetBasicAuth("wrong-username", "wrong-password")

				return req
			},
			expectedError: ErrAuthHeader,
		},
		"empty_auth_header": {
			opts: &BasicAuthConfig{
				Username: "convoy-ingester",
				Password: "convoy-password",
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T, c *BasicAuthConfig) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				return req
			},
			expectedError: ErrAuthHeaderCannotBeEmpty,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			v := BasicAuthVerifier{tc.opts}
			req := tc.requestFn(t, tc.opts)

			// Assert
			err := v.VerifyRequest(req, tc.payload)

			// Act.
			require.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func Test_APIKeyVerifier_VerifyRequest(t *testing.T) {
	tests := map[string]struct {
		opts          *APIKeyConfig
		payload       []byte
		requestFn     func(t *testing.T, c *APIKeyConfig) *http.Request
		expectedError error
	}{
		"invalid_api_key": {
			opts: &APIKeyConfig{
				APIKey: "sec_apikeysecret",
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T, c *APIKeyConfig) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				req.Header.Add("Authorization", "sec_invalidkey")
				return req
			},
			expectedError: ErrAuthHeader,
		},
		"valid_request": {
			opts: &APIKeyConfig{
				APIKey: "sec_apikeysecret",
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T, c *APIKeyConfig) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				req.Header.Add("Authorization", "sec_apikeysecret")
				return req
			},
			expectedError: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange
			v := APIKeyVerifier{tc.opts}
			req := tc.requestFn(t, tc.opts)

			// Assert
			err := v.VerifyRequest(req, tc.payload)

			// Act.
			require.ErrorIs(t, err, tc.expectedError)
		})
	}
}
