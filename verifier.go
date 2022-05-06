package ingester

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"hash"
	"net"
	"net/http"
	"strings"
)

var ErrAlgoNotFound = errors.New("Algorithm not found")
var ErrInvalidIP = errors.New("Source IP not supported")
var ErrCannotReadRequestBody = errors.New("Failed to read request body")
var ErrHashDoesNotMatch = errors.New("Invalid Signature - Hash does not match")
var ErrCannotDecodeMACHeader = errors.New("Cannot decode MAC header")
var ErrSignatureCannotBeEmpty = errors.New("Signature cannot be empty")
var ErrAuthHeader = errors.New("Invalid Authorization header")
var ErrAuthHeaderCannotBeEmpty = errors.New("Auth header cannot be empty")
var ErrInvalidHeaderStructure = errors.New("Invalid header structure")
var ErrInvalidAuthLength = errors.New("Invalid Basic Auth Length")

type Verifier interface {
	VerifyRequest(r *http.Request, payload []byte) error
}

type HmacVerifier struct {
	config *HmacConfig
}

func (hV *HmacVerifier) VerifyRequest(r *http.Request, payload []byte) error {
	if hV.config == nil {
		return errors.New("Verifier Config Error")
	}

	// Check Signature.
	hash, err := hV.getHashFunction(hV.config.Hash)
	if err != nil {
		return err
	}

	rHeader := r.Header.Get(hV.config.Header)

	if len(strings.TrimSpace(rHeader)) == 0 {
		return ErrSignatureCannotBeEmpty
	}

	mac := hmac.New(hash, []byte(hV.config.Secret))
	mac.Write(payload)
	eMAC := mac.Sum(nil)
	sMAC, err := hex.DecodeString(rHeader)
	if err != nil {
		return ErrCannotDecodeMACHeader
	}

	validMAC := hmac.Equal(sMAC, eMAC)
	if !validMAC {
		return ErrHashDoesNotMatch
	}

	return nil
}

func (hV *HmacVerifier) getHashFunction(algo string) (func() hash.Hash, error) {
	switch algo {
	case "SHA256":
		return sha256.New, nil
	case "SHA512":
		return sha512.New, nil
	default:
		return nil, ErrAlgoNotFound
	}
}

type BasicAuthVerifier struct {
	config *BasicAuthConfig
}

func (baV *BasicAuthVerifier) VerifyRequest(r *http.Request, payload []byte) error {
	val := r.Header.Get("Authorization")
	authInfo := strings.Split(val, " ")

	if len(authInfo) != 2 {
		return ErrInvalidHeaderStructure
	}

	credentials, err := base64.StdEncoding.DecodeString(authInfo[1])
	if err != nil {
		return ErrInvalidHeaderStructure
	}

	creds := strings.Split(string(credentials), ":")

	if len(creds) != 2 {
		return ErrInvalidAuthLength
	}

	if creds[0] != baV.config.Username && creds[1] != baV.config.Password {
		return ErrAuthHeader
	}

	return nil
}

type APIKeyVerifier struct {
	config *APIKeyConfig
}

func (aV *APIKeyVerifier) VerifyRequest(r *http.Request, payload []byte) error {
	authHeader := "Authorization"

	if len(strings.TrimSpace(aV.config.Header)) != 0 {
		authHeader = aV.config.Header
		val := r.Header.Get(authHeader)

		if len(strings.TrimSpace(val)) == 0 {
			return ErrAuthHeader
		}

		if val != aV.config.APIKey {
			return ErrAuthHeader
		}
	}

	val := r.Header.Get(authHeader)
	authInfo := strings.Split(val, " ")

	if len(authInfo) != 2 {
		return ErrInvalidHeaderStructure
	}

	if authInfo[1] != aV.config.APIKey {
		return ErrAuthHeader
	}

	return nil
}

type IPAddressVerifier struct {
	config *IPAddressConfig
}

func (ipV *IPAddressVerifier) VerifyRequest(r *http.Request, payload []byte) error {
	return nil
}

// Verify IP Address.
// See here - https://husobee.github.io/golang/ip-address/2015/12/17/remote-ip-go.html
func getIPAddress(r *http.Request) string {
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		addresses := strings.Split(r.Header.Get(h), ",")
		// march from right to left until we get a public address
		// that will be the address right before our proxy.
		for i := len(addresses) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(addresses[i])
			// header can contain spaces too, strip those out.
			realIP := net.ParseIP(ip)
			if !realIP.IsGlobalUnicast() || isPrivateSubnet(realIP) {
				// bad address, go to next
				continue
			}
			return ip
		}
	}
	return ""
}

//ipRange - a structure that holds the start and end of a range of ip addresses
type ipRange struct {
	start net.IP
	end   net.IP
}

// inRange - check to see if a given ip address is within a range given
func inRange(r ipRange, ipAddress net.IP) bool {
	// strcmp type byte comparison
	if bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0 {
		return true
	}
	return false
}

var privateRanges = []ipRange{
	ipRange{
		start: net.ParseIP("10.0.0.0"),
		end:   net.ParseIP("10.255.255.255"),
	},
	ipRange{
		start: net.ParseIP("100.64.0.0"),
		end:   net.ParseIP("100.127.255.255"),
	},
	ipRange{
		start: net.ParseIP("172.16.0.0"),
		end:   net.ParseIP("172.31.255.255"),
	},
	ipRange{
		start: net.ParseIP("192.0.0.0"),
		end:   net.ParseIP("192.0.0.255"),
	},
	ipRange{
		start: net.ParseIP("192.168.0.0"),
		end:   net.ParseIP("192.168.255.255"),
	},
	ipRange{
		start: net.ParseIP("198.18.0.0"),
		end:   net.ParseIP("198.19.255.255"),
	},
}

// isPrivateSubnet - check to see if this ip is in a private subnet
func isPrivateSubnet(ipAddress net.IP) bool {
	// my use case is only concerned with ipv4 atm
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		// iterate over all our ranges
		for _, r := range privateRanges {
			// check if this ip is in a private range
			if inRange(r, ipAddress) {
				return true
			}
		}
	}
	return false
}
