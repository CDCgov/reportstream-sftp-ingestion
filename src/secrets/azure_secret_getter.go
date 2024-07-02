package secrets

import (
	"context"
	"crypto/rsa"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"os"
	"time"
)

type SecretGetter struct {
	client *azsecrets.Client
}

func NewSecretGetter() (SecretGetter, error) {

	// Create a credential using the NewDefaultAzureCredential type.
	cred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
		ClientOptions: policy.ClientOptions{
			Retry: policy.RetryOptions{
				TryTimeout: 60 * time.Second,
			},
		},
	})

	if err != nil {
		slog.Error("failed to obtain a credential: ", slog.Any(utils.ErrorKey, err))
		return SecretGetter{}, err
	}

	vaultURI := os.Getenv("AZURE_KEY_VAULT_URI")

	// Establish a connection to the Key Vault client
	newClient, err := azsecrets.NewClient(vaultURI, cred, nil)
	if err != nil {
		slog.Error("failed to create a client: ", slog.Any(utils.ErrorKey, err))
		return SecretGetter{}, err
	}

	return SecretGetter{client: newClient}, nil
}

func (credentialGetter SecretGetter) GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error) {
	slog.Info("Reading private key from Azure", slog.String("name", privateKeyName))

	pem, err := credentialGetter.GetSecret(privateKeyName)
	if err != nil {
		return nil, err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pem))
	if err != nil {
		return nil, err
	}

	slog.Info("parsed pem to key")

	return key, nil
}

func (credentialGetter SecretGetter) GetSecret(secretName string) (string, error) {
	secretResponse, err := credentialGetter.client.GetSecret(context.TODO(), secretName, "", nil)
	if err != nil {
		slog.Error("failed to get the secret ", slog.Any(utils.ErrorKey, err))
		return "", err
	}
	return *secretResponse.Secret.Value, err
}
