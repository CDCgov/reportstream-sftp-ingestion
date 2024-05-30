package utils

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type SecretGetter interface {
	GetSecret(context context.Context, secretKey string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
}
