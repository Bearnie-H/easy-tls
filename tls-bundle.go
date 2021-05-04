package easytls

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

// KeyPair is a single matched pair of TLS Certificate and Key files.
type KeyPair struct {
	Certificate string
	Key         string
}

// TLSBundle is a toggle-able set of TLS resources to be used to generate a
// valid tls.Config struct, to be used with the http package.  This is composed
// of a whitelisted set of Certificate Authorities, a TLS Certificate and Key
// to use, a Peer Authentication policy, and a toggle to turn on or off the TLS
// settings.
type TLSBundle struct {

	// AuthorityCertificates is a set of filenames, of the set of Certificate
	// Authorities to use when building the whitelist of acceptable Certificate
	// Authorities for TLS communications.
	//
	// If no explicit CA Certificates are provided, the default set of System
	// Certificate Authorities will be used.
	AuthorityCertificates []string

	// KeyPair is the matched pair of "Client" Certificate and Key to use and
	// present during a TLS handshake.
	//
	// Optional for Clients unless the server being contacted is configured to
	// require client certificates.
	//
	// Required for Servers.
	KeyPair KeyPair

	// Auth defines the policy to use during the TLS handshake to verify the
	// other host's certificate.
	Auth tls.ClientAuthType

	// Enabled allows this to be toggled. If disabled, this will create an
	// nil tls.Config when used, turning TLS off for whatever client or server
	// which uses the returned tls.Config{}.
	Enabled bool
}

// NewTLSBundle will create and return a new default TLSBundle from a given set of resources.
// This will be enabled by default, with an ClientAuthType which doesn't require or care about client
// certificates.
func NewTLSBundle(CertificateFile, KeyFile string, CertificateAuthorityFiles ...string) *TLSBundle {
	return &TLSBundle{
		AuthorityCertificates: CertificateAuthorityFiles,
		KeyPair: KeyPair{
			Certificate: CertificateFile,
			Key:         KeyFile,
		},
		Auth:    tls.NoClientCert,
		Enabled: true,
	}
}

// NewTLSConfig will convert the TLSBundle, containing the filenames of the
// relevant certificates and Authorization policy, into a workable tls.Config
// object, ready to be used by either a SimpleClient or SimpleServer application.
func NewTLSConfig(TLS *TLSBundle) (*tls.Config, error) {

	// If no TLS bundle is provided, or is it marked disabled, don't set any TLS settings.
	if TLS == nil || !TLS.Enabled {
		return nil, nil
	}

	// Create the tls.Config to return if everything goes well.
	returnConfig := &tls.Config{}

	// If no KeyPairs are provided, don't attempt to load Client-side certificates
	if TLS.KeyPair.Certificate != "" && TLS.KeyPair.Key != "" {
		cert, err := tls.LoadX509KeyPair(TLS.KeyPair.Certificate, TLS.KeyPair.Key)
		if err != nil {
			return nil, err
		}
		returnConfig.Certificates = append(returnConfig.Certificates, cert)
	}

	// If CA Certificates are provided, attempt to load and build a pool from the full set
	if TLS.AuthorityCertificates == nil || len(TLS.AuthorityCertificates) > 0 {
		caCertPool := x509.NewCertPool()

		for _, AuthorityCert := range TLS.AuthorityCertificates {

			// Load the CA cert
			caCert, err := ioutil.ReadFile(AuthorityCert)
			if err != nil {
				return nil, err
			}

			// Create and append the CA Cert to the pool of approved certificate authorities.
			// This sets up so that ONLY the CA who signed this certificate can verify the received server certificate.
			caCertPool.AppendCertsFromPEM(caCert)
		}

		// The way we implement TLS CAs here expects that the full set of accepted CAs is a whitelist, and whether we check or care about certificates is based on the ClientAuth.
		returnConfig.RootCAs = caCertPool
		returnConfig.ClientCAs = caCertPool
	} else {
		// If no CA Certificates are provided, default to the system Certificate Pool
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}

		returnConfig.RootCAs = caCertPool
		returnConfig.ClientCAs = caCertPool
	}

	// If there are Certificates, the TLS min version can be set.
	// Default to the maximum supported version, sorry if this breaks old applications.
	returnConfig.MinVersion = tls.VersionTLS13

	// Define how the Client Certificates will be checked.
	returnConfig.ClientAuth = TLS.Auth

	return returnConfig, nil
}
