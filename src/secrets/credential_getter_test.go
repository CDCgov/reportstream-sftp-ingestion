package secrets

import (
	"github.com/CDCgov/reportstream-sftp-ingestion/utils"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
)

func Test_GetCredentialGetter_EnvIsNotSet_ReturnsLocalCredentialGetter(t *testing.T) {
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	os.Setenv("ENV", "")
	defer os.Unsetenv("ENV")

	credentialGetter, err := GetCredentialGetter()

	assert.NotNil(t, credentialGetter)
	assert.Contains(t, buffer.String(), "Using local credentials")
	assert.NoError(t, err)
}

func Test_GetCredentialGetter_EnvIsLocal_ReturnsLocalCredentialGetter(t *testing.T) {
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	os.Setenv("ENV", "local")
	defer os.Unsetenv("ENV")

	credentialGetter, err := GetCredentialGetter()

	assert.NotNil(t, credentialGetter)
	assert.Contains(t, buffer.String(), "Using local credentials")
	assert.NoError(t, err)
}

func Test_GetCredentialGetter_EnvIsNotLocal_ReturnsAzureCredentialGetter(t *testing.T) {
	buffer, defaultLogger := utils.SetupLogger()
	defer slog.SetDefault(defaultLogger)

	os.Setenv("ENV", "not local")
	defer os.Unsetenv("ENV")

	credentialGetter, err := GetCredentialGetter()

	assert.NotNil(t, credentialGetter)
	assert.Contains(t, buffer.String(), "Using Azure credentials")
	assert.NoError(t, err)
}
