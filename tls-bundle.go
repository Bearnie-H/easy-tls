package easytls

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
)

// KeyPair represents the filenames to a pair of matching TLS Key/Certificate files.
type KeyPair struct {
	Certificate string
	Key         string
}

// TLSBundle represents the set of TLS information required by Dune to assert 2-way TLS verification.
type TLSBundle struct {
	AuthorityCertificates []string
	KeyPair               KeyPair
	Auth                  tls.ClientAuthType
	Enabled               bool `json:"-"`
}

// NewTLSConfig will convert the TLSBundle, containing the filenames of the relevant certificates and Authorization policy, into a workable tls.Config object, ready to be used by either a SimpleClient or SimpleServer application.
func NewTLSConfig(TLS *TLSBundle) (*tls.Config, error) {

	returnConfig := &tls.Config{}

	// If no TLS bundle is provided, or is it marked invalid, don't set any TLS settings.
	if TLS == nil {
		return &tls.Config{}, nil
	}

	if !TLS.Enabled {
		return &tls.Config{}, nil
	}

	// If no KeyPairs are provided, don't attempt to load Client-side certificates
	if TLS.KeyPair.Certificate == "" || TLS.KeyPair.Key == "" {
		cert, err := tls.LoadX509KeyPair(TLS.KeyPair.Certificate, TLS.KeyPair.Key)
		if err != nil {
			log.Printf("Failed to load  certificate - %s\n", err)
			return &tls.Config{}, err
		}
		returnConfig.Certificates = append(returnConfig.Certificates, cert)

		// If there are Certificates, the TLS min version can be set (1.2 is used here for backwards-compatability)
		returnConfig.MinVersion = tls.VersionTLS12
		TLS.Enabled = true
	}

	// If no CA Certificates are provided, don't attempt to load any
	if len(TLS.AuthorityCertificates) != 0 {
		caCertPool := x509.NewCertPool()
		for _, AutorityCert := range TLS.AuthorityCertificates {
			// Load the CA cert
			caCert, err := ioutil.ReadFile(AutorityCert)
			if err != nil {
				log.Printf("Failed to load CA certificate - %s\n", err)
				return &tls.Config{}, err
			}
			// Create and append the CA Cert to the pool of approved certificate authorities.
			// This sets up so that ONLY the CA who signed this 's certificate can verify the recieved server certificate.
			caCertPool.AppendCertsFromPEM(caCert)

			// The way we implement TLS CAs here expects that the full set of accepted CAs is a whitelist, and whether we check or care about certificates is based on the ClientAuth.
			returnConfig.RootCAs = caCertPool
			returnConfig.ClientCAs = caCertPool
		}

		// Set this, if it wasn't set before, as there are now CA certs.
		returnConfig.MinVersion = tls.VersionTLS12
		TLS.Enabled = true
	}

	// Define how the Client Certificates will be checked.
	returnConfig.ClientAuth = TLS.Auth

	return returnConfig, nil
}
