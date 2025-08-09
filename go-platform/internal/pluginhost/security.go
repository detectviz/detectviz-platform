package pluginhost

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// LoadTLSConfig 由檔案路徑載入 mTLS 憑證。若任一為空，回傳 nil 代表使用明文（僅限開發）。
func LoadTLSConfig(certPath, keyPath, caPath string) (*tls.Config, error) {
	if certPath == "" || keyPath == "" {
		return nil, nil
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	var pool *x509.CertPool
	if caPath != "" {
		bs, err := os.ReadFile(caPath)
		if err != nil {
			return nil, err
		}
		pool = x509.NewCertPool()
		if !pool.AppendCertsFromPEM(bs) {
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