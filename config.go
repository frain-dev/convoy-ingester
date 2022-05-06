package ingester

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type Configuration []ProviderConfig

type ProviderConfig struct {
	Name           string         `json:"name"`
	AppID          string         `json:"app_id"`
	VerifierConfig VerifierConfig `json:"verifier_config"`
}

type VerifierConfig struct {
	*HmacConfig
	*BasicAuthConfig
	*APIKeyConfig
	*IPAddressConfig
}

type HmacConfig struct {
	Header string `json:"header"`
	Hash   string `json:"hash"`
	Secret string `json:"secret"`
}

type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type APIKeyConfig struct {
	Header string `json:"header"`
	APIKey string `json:"api_key"`
}

type IPAddressConfig struct {
	IPSafelist []string `json:"ip_safelist"`
}

func (vC *VerifierConfig) UnmarshalJSON(data []byte) error {
	temp := struct {
		Type string `json:"type"`
	}{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	switch temp.Type {
	case "hmac":
		var c HmacConfig
		if err := json.Unmarshal(data, &c); err != nil {
			return err
		}

		// TODO(subomi): Invalidate all other verifiers.
		vC.HmacConfig = &c
		return nil
	case "api_key":
		var c APIKeyConfig
		if err := json.Unmarshal(data, &c); err != nil {
			return err
		}

		vC.APIKeyConfig = &c
		return nil
	case "basic_auth":
		var c BasicAuthConfig
		if err := json.Unmarshal(data, &c); err != nil {
			return err
		}

		vC.BasicAuthConfig = &c
		return nil
	case "ip_address":
		var c IPAddressConfig
		if err := json.Unmarshal(data, &c); err != nil {
			return err
		}

		vC.IPAddressConfig = &c
		return nil
	default:
		//TODO(subomi): rewrite this to an error type
		return errors.New("Invalid verification config")
	}
}

func LoadConfig(env string) error {
	f := os.Getenv(env)
	if len(strings.TrimSpace(f)) == 0 {
		return errors.New("Configuration cannot be empty")
	}

	s := strings.NewReader(f)
	if err := json.NewDecoder(s).Decode(&configStore); err != nil {
		return err
	}

	return nil
}
