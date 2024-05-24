package local

import (
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"path/filepath"
)

type CredentialGetter struct {
}

func (credentialGetter CredentialGetter) GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error) {
	//TODO - get this from environment variables?
	//- have e.g. a credential getter interface that varies by 'local or not'
	// put interface in this package, one implementation in this package, another implementation in local

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
