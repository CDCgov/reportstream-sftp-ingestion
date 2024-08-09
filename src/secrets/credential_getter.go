package secrets

import (
	"crypto/rsa"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"log/slog"
)

// The CredentialGetter interface is about getting private keys
// Currently we can get credentials from an Azure vault in deployed envs or from the local file system
type CredentialGetter interface {
	GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error)
	GetSecret(secretName string) (string, error)
}

func GetCredentialGetter() (CredentialGetter, error) {
	var credentialGetter CredentialGetter

	if utils.EnvironmentName() == "local" {
		slog.Info("Using local credentials")
		credentialGetter = LocalCredentialGetter{}
	} else {
		slog.Info("Using Azure credentials")
		var err error
		credentialGetter, err = NewSecretGetter()
		if err != nil {
			return nil, err
		}
	}
	return credentialGetter, nil
}
