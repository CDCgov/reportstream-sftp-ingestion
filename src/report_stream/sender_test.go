package report_stream

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func Test_New_Sender(t *testing.T) {
	setUpTest()
	defer tearDownTest()

	sender := NewSender()

	assert.Equal(t, os.Getenv("REPORT_STREAM_URL_PREFIX"), sender.BaseUrl)
	assert.Equal(t, os.Getenv("FLEXION_PRIVATE_KEY_NAME"), sender.PrivateKeyName)
	assert.Equal(t, os.Getenv("FLEXION_CLIENT_NAME"), sender.ClientName)
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
