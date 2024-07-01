package utils

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
)

func Test_GetCredentialGetter_returnLocalCredentialGetterWhenEnvNotSet(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	os.Setenv("ENV", "")
	defer os.Unsetenv("ENV")

	credentialGetter, err := GetCredentialGetter()

	assert.NotNil(t, credentialGetter)
	assert.Contains(t, buffer.String(), "Using local credentials")
	assert.NoError(t, err)
}

func Test_GetCredentialGetter_returnLocalCredentialGetterWhenEnvSet(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	os.Setenv("ENV", "local")
	defer os.Unsetenv("ENV")

	credentialGetter, err := GetCredentialGetter()

	assert.NotNil(t, credentialGetter)
	assert.Contains(t, buffer.String(), "Using local credentials")
	assert.NoError(t, err)
}

func Test_GetCredentialGetter_returnAzureCredentialGetter(t *testing.T) {
	defaultLogger := slog.Default()
	defer slog.SetDefault(defaultLogger)

	buffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, nil)))

	os.Setenv("ENV", "not local")
	defer os.Unsetenv("ENV")

	credentialGetter, err := GetCredentialGetter()

	assert.NotNil(t, credentialGetter)
	assert.Contains(t, buffer.String(), "Using Azure credentials")
	assert.NoError(t, err)
}
