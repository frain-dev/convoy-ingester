package ingester

import (
	"net/http"
)

// Store
type ProviderStore map[string]*Provider

type Provider struct {
	Name     string
	AppID    string
	verifier Verifier
}

func (p *Provider) VerifyRequest(r *http.Request, payload []byte) error {
	return p.verifier.VerifyRequest(r, payload)
}

func LoadProviderStore() error {

	// Create registry from configuration
	for _, c := range *configStore {
		p := &Provider{
			Name:  c.Name,
			AppID: c.AppID,
		}

		if c.VerifierConfig.HmacConfig != nil {
			p.verifier = &HmacVerifier{c.VerifierConfig.HmacConfig}
		} else if c.VerifierConfig.BasicAuthConfig != nil {
			p.verifier = &BasicAuthVerifier{c.VerifierConfig.BasicAuthConfig}
		} else if c.VerifierConfig.APIKeyConfig != nil {
			p.verifier = &APIKeyVerifier{c.VerifierConfig.APIKeyConfig}
		}

		providerStore[c.Name] = p
	}

	return nil
}
