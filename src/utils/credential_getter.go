package utils

import (
	"crypto/rsa"
	"github.com/CDCgov/reportstream-sftp-ingestion/secrets"
	"log/slog"
	"os"
)

// The CredentialGetter interface is about getting private keys
// Currently we can get credentials from an Azure vault in deployed envs or from the local file system
type CredentialGetter interface {
	GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error)
	GetSecret(secretName string) (string, error)
}

func GetCredentialGetter() (CredentialGetter, error) {
	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "local"
	}

	var credentialGetter CredentialGetter

	if environment == "local" {
		slog.Info("Using local credentials")
		credentialGetter = secrets.CredentialGetter{}
	} else {
		slog.Info("Using Azure credentials")
		var err error
		credentialGetter, err = secrets.NewSecretGetter()
		if err != nil {
			return nil, err
		}
	}
	return credentialGetter, nil
}
