package verifier

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

func Test_VerifyRequest(t *testing.T) {
	tests := map[string]struct {
		opts          *VerifierOptions
		payload       []byte
		requestFn     func(t *testing.T) *http.Request
		expected      bool
		expectedError error
	}{
		"invalid_signature": {
			opts: &VerifierOptions{
				Header:      "X-Convoy-Signature",
				Hash:        "SHA512",
				Secret:      "Convoy",
				IPWhitelist: []string{},
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				hash := hex.EncodeToString([]byte(`Obviously wrong hash`))

				req.Header.Add("X-Convoy-Signature", hash)
				return req
			},
			expected:      false,
			expectedError: ErrHashDoesNotMatch,
		},
		"invalid_hex": {
			opts: &VerifierOptions{
				Header:      "X-Convoy-Signature",
				Hash:        "SHA512",
				Secret:      "Convoy",
				IPWhitelist: []string{},
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				hash := "Hash with characters outside hex"

				req.Header.Add("X-Convoy-Signature", hash)
				return req
			},
			expected:      false,
			expectedError: ErrCannotDecodeMACHeader,
		},
		"empty_signature": {
			opts: &VerifierOptions{
				Header:      "X-Convoy-Signature",
				Hash:        "SHA512",
				Secret:      "Convoy",
				IPWhitelist: []string{},
			},
			payload: []byte(`Test Payload Body`),
			requestFn: func(t *testing.T) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				req.Header.Add("X-Convoy-Signature", "")
				return req
			},
			expected:      false,
			expectedError: ErrSignatureCannotBeEmpty,
		},
		"invalid_ip": {
			opts: &VerifierOptions{
				Header:      "X-Convoy-Signature",
				Hash:        "SHA512",
				Secret:      "Convoy",
				IPWhitelist: []string{"52.52.52.52"},
			},
			requestFn: func(t *testing.T) *http.Request {
				req, err := http.NewRequest("POST", "URL", strings.NewReader(``))
				require.NoError(t, err)

				req.Header.Add("X-Forwarded-For", "50.50.50.50")
				return req
			},
			expected:      false,
			expectedError: ErrInvalidIP,
		},
		"valid_request": {
			opts: &VerifierOptions{
				Header:      "X-Convoy-Signature",
				Hash:        "SHA512",
				Secret:      "Convoy",
				IPWhitelist: []string{"52.52.52.52"},
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
			expected:      true,
			expectedError: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Arrange.
			v := NewVerifier(tc.opts)
			req := tc.requestFn(t)

			// Assert.
			ok, err := v.VerifyRequest(req, tc.payload)

			// Act.
			require.ErrorIs(t, err, tc.expectedError)
			require.Equal(t, ok, tc.expected)
		})
	}
}
