package report_stream

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

type MockCredentialGetter struct {
	mock.Mock
}

func (m *MockCredentialGetter) GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error) {

	args := m.Called(privateKeyName)

	return args.Get(0).(*rsa.PrivateKey), args.Error(1)

}

func Test_Sender_NewSender_InitsWithValues(t *testing.T) {
	setUpTest()
	defer tearDownTest()

	sender := NewSender()

	assert.Equal(t, os.Getenv("REPORT_STREAM_URL_PREFIX"), sender.BaseUrl)
	assert.Equal(t, os.Getenv("FLEXION_PRIVATE_KEY_NAME"), sender.PrivateKeyName)
	assert.Equal(t, os.Getenv("FLEXION_CLIENT_NAME"), sender.ClientName)
}

func Test_Sender_GenerateJWT_ReturnsJWT(t *testing.T) {
	setUpTest()
	defer tearDownTest()

	sender := NewSender()
	mockCredentialGetter := new(MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, nil)
	jwt, err := sender.GenerateJwt()

	assert.NoError(t, err)
	assert.NotEmpty(t, jwt)
}

func Test_Sender_GenerateJWT_ReturnsError_WhenGetPrivateKeyReturnsError(t *testing.T) {
	setUpTest()
	defer tearDownTest()

	sender := NewSender()
	mockCredentialGetter := new(MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	mockCredentialGetter.On("GetPrivateKey", "key").Return(&rsa.PrivateKey{}, errors.New("failed to retrieve private key"))
	_, err := sender.GenerateJwt()

	assert.Error(t, err)
}

func setUpTest() {
	os.Setenv("ENV", "local")
	os.Setenv("REPORT_STREAM_URL_PREFIX", "rs.com")
	os.Setenv("FLEXION_PRIVATE_KEY_NAME", "key")
	os.Setenv("FLEXION_CLIENT_NAME", "client")
}

func tearDownTest() {
	os.Unsetenv("ENV")
	os.Unsetenv("REPORT_STREAM_URL_PREFIX")
	os.Unsetenv("FLEXION_PRIVATE_KEY_NAME")
	os.Unsetenv("FLEXION_CLIENT_NAME")
}
