package ingester

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"hash"
	"io/ioutil"
	"net/http"
)

var ErrAlgoNotFound = errors.New("Algorithm not found")
var ErrInvalidIP = errors.New("Source IP not supported")
var ErrCannotReadRequestBody = errors.New("Failed to read request body")
var ErrHashDoesNotMatch = errors.New("Invalid Signature - Hash does not match")

// VerifierOptions
type VerifierOptions struct {
	Header      string
	Hash        string
	Secret      string
	IPWhitelist []string
}

type Verifier struct {
	opts *VerifierOptions
}

func NewVerifier(opts *VerifierOptions) *Verifier {
	return &Verifier{
		opts: opts,
	}
}

func (v *Verifier) VerifyRequest(r *http.Request) (bool, error) {
	// Check IP Address.
	var validIP bool
	ipAddr := r.RemoteAddr

	validIP = false
	for _, v := range v.opts.IPWhitelist {
		if ipAddr == v {
			validIP = true
			break
		}
	}

	if !validIP {
		return false, ErrInvalidIP
	}

	// Check Signature.
	hash, err := v.getHashFunction(v.opts.Hash)
	if err != nil {
		return false, err
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return false, ErrCannotReadRequestBody
	}

	mac := hmac.New(hash, []byte(v.opts.Secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	sentMAC := r.Header.Get(v.opts.Header)

	validMAC := hmac.Equal([]byte(sentMAC), expectedMAC)

	if !validMAC {
		return false, ErrHashDoesNotMatch
	}

	return true, nil
}

func (v *Verifier) getHashFunction(algorithm string) (func() hash.Hash, error) {
	switch algorithm {
	case "SHA256":
		return sha256.New, nil
	case "SHA512":
		return sha512.New, nil
	default:
		return nil, ErrAlgoNotFound
	}
}
