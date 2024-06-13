package senders

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type SenderTestSuite struct {
	suite.Suite
}

func (suite *SenderTestSuite) SetupTest() {
	os.Setenv("ENV", "local")
	os.Setenv("REPORT_STREAM_URL_PREFIX", "rs.com")
	os.Setenv("FLEXION_PRIVATE_KEY_NAME", "key")
	os.Setenv("FLEXION_CLIENT_NAME", "client")
}

func (suite *SenderTestSuite) TearDownTest() {
	os.Unsetenv("ENV")
	os.Unsetenv("REPORT_STREAM_URL_PREFIX")
	os.Unsetenv("FLEXION_PRIVATE_KEY_NAME")
	os.Unsetenv("FLEXION_CLIENT_NAME")
}

type MockCredentialGetter struct {
	mock.Mock
}

func (m *MockCredentialGetter) GetPrivateKey(privateKeyName string) (*rsa.PrivateKey, error) {
	args := m.Called(privateKeyName)
	return args.Get(0).(*rsa.PrivateKey), args.Error(1)
}

func (suite *SenderTestSuite) Test_Sender_NewSender_InitsWithValues() {

	sender, err := NewSender()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), os.Getenv("REPORT_STREAM_URL_PREFIX"), sender.baseUrl)
	assert.Equal(suite.T(), os.Getenv("FLEXION_PRIVATE_KEY_NAME"), sender.privateKeyName)
	assert.Equal(suite.T(), os.Getenv("FLEXION_CLIENT_NAME"), sender.clientName)
}

func (suite *SenderTestSuite) Test_Sender_GenerateJWT_ReturnsJWT() {

	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	testKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(suite.T(), err)

	mockCredentialGetter.On("GetPrivateKey", "key").Return(testKey, nil)
	jwt, err := sender.generateJwt()

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), jwt)
}

func (suite *SenderTestSuite) Test_Sender_GenerateJWT_ReturnsError_WhenGetPrivateKeyReturnsError() {

	sender, err := NewSender()
	assert.NoError(suite.T(), err)

	mockCredentialGetter := new(MockCredentialGetter)
	sender.credentialGetter = mockCredentialGetter

	mockCredentialGetter.On("GetPrivateKey", "key").Return(&rsa.PrivateKey{}, errors.New("failed to retrieve private key"))
	_, err = sender.generateJwt()

	assert.Error(suite.T(), err)
}

func Test_SenderTestSuite(t *testing.T) {
	suite.Run(t, new(SenderTestSuite))
}
