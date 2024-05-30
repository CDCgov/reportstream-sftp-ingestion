package azure

import (
	"context"
	"crypto/rsa"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"os"
	"time"
)

type CredentialGetter struct{}

func (credentialGetter CredentialGetter) GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error) {
	vaultURI := os.Getenv("AZURE_KEY_VAULT_URI")

	// Create a credential using the NewDefaultAzureCredential type.
	cred, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
		ClientOptions: policy.ClientOptions{
			Retry: policy.RetryOptions{
				TryTimeout: 60 * time.Second,
			},
		},
	})

	if err != nil {
		slog.Error("failed to obtain a credential: ", slog.Any("error", err))
		return nil, err
	}

	// Establish a connection to the Key Vault client
	client, err := azsecrets.NewClient(vaultURI, cred, nil)
	if err != nil {
		slog.Error("failed to create a client: ", slog.Any("error", err))
		return nil, err
	}

	version := ""
	resp, err := client.GetSecret(context.TODO(), privateKeyName, version, nil)
	if err != nil {
		slog.Error("failed to get the secret ", slog.Any("error", err))
		return nil, err
	}

	pem := *resp.Value

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pem))
	if err != nil {
		return nil, err
	}
	slog.Info("parsed pem to key")
	return key, nil
}
