package pluginhost

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// LoadTLSConfig 由憑證內容載入 mTLS 設定。若任一為空，回傳 nil 代表使用明文（僅限開發）。
func LoadTLSConfig(certData, keyData, caData []byte) (*tls.Config, error) {
	if len(certData) == 0 || len(keyData) == 0 {
		return nil, nil // No cert/key provided, assuming insecure.
	}

	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to load x509 key pair: %w", err)
	}

	var pool *x509.CertPool
	if len(caData) > 0 {
		pool = x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("failed to append CA PEM")
		}
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
		MinVersion:   tls.VersionTLS13,
	}, nil
}
