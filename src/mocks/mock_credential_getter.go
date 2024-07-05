package mocks

import (
	"crypto/rsa"
	"github.com/stretchr/testify/mock"
)

type MockCredentialGetter struct {
	mock.Mock
}

func (m *MockCredentialGetter) GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error) {
	args := m.Called(privateKeyName)
	return args.Get(0).(*rsa.PrivateKey), args.Error(1)
}
func (m *MockCredentialGetter) GetSecret(secretName string) (string, error) {
	args := m.Called(secretName)
	return args.Get(0).(string), args.Error(1)
}
