package utils

import "crypto/rsa"

// The CredentialGetter interface is about getting private keys
// Currently we can get credentials from an Azure vault in deployed envs or from the local file system
type CredentialGetter interface {
	GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error)
}
