package secrets

import (
	"crypto/rsa"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"os"
	"path/filepath"
)

type LocalCredentialGetter struct {
}

func (credentialGetter LocalCredentialGetter) GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error) {
	slog.Info("Reading private key from local hard drive", slog.String("name", privateKeyName))

	pem, err := credentialGetter.GetSecret(privateKeyName)
	if err != nil {
		return nil, err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pem))
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (credentialGetter LocalCredentialGetter) GetSecret(secretName string) (string, error) {
	slog.Info("Reading secret from local hard drive", slog.String("name", secretName))

	secret, err := os.ReadFile(filepath.Join("mock_credentials", secretName))
	if err != nil {
		return "", err
	}

	return string(secret), nil
}
