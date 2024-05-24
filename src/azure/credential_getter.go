package azure

import (
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"path/filepath"
)

type CredentialGetter struct{}

func (credentialGetter CredentialGetter) GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error) {
	//TODO - Put the real code here

	pem, err := os.ReadFile(filepath.Join("mock_credentials", fmt.Sprintf("%s.pem", privateKeyName)))
	if err != nil {
		return nil, err
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(pem)
	if err != nil {
		return nil, err
	}
	return key, nil
}
