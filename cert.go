package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"math/big"
	"time"
)

// method creates a self-signed TLS configuration for the QUIC server
// used to secure QUIC connections using TLS 1.3.
// certificate is generated on the fly and is not trusted by browsers - only for local testing
func generateTLSConfig() *tls.Config {
	// Generate a new RSA private key (2048-bit)
	key, _ := rsa.GenerateKey(rand.Reader, 2048)

	// create a self-signed certificate template
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),                   // Arbitrary serial number
		NotBefore:             time.Now(),                      // Certificate validity start time
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// self-sign the certificate using the generated private key
	certDER, _ := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)

	// create a tls.Certificate using the cert and private key
	keyPEM := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}

	// return a TLS config with the generated certificate and the ALPN protocol identifier
	return &tls.Config{
		Certificates: []tls.Certificate{keyPEM},
		NextProtos:   []string{"qtgp-demo"}, // ALPN: used by QUIC to negotiate the application protocol
	}
}
