package azure

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

type MockSecretsGetter struct {
	mock.Mock
}

func (m *MockSecretsGetter) GetSecret(context context.Context, secretKey string, version string, _ *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	args := m.Called(context, secretKey, version, nil)

	return args.Get(0).(azsecrets.GetSecretResponse), args.Error(1)
}

func Test_Credential_Getter_GetsPrivateKey(t *testing.T) {

	os.Setenv("AZURE_KEY_VAULT_URI", "https://mockTests.com")
	mockSecretGetter := new(MockSecretsGetter)
	secret := NewSecretGetter()

	secret.secretGetter = mockSecretGetter

	//TODO: do what the below says
	secretValue := "PUT SOME REAL KEY HERE"
	response := azsecrets.GetSecretResponse{
		Secret: azsecrets.Secret{Value: &secretValue},
	}

	mockSecretGetter.On("GetSecret", context.TODO(), "testKey", "", nil).Return(response, nil)

	key, err := secret.GetPrivateKey("testKey")

	assert.NoError(t, err)
	assert.NotEmpty(t, key)
}
