module github.com/CDCgov/reportstream-sftp-ingestion

go 1.23

toolchain go1.23.4

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.17.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.8.1
	github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/azeventgrid v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets v1.3.0
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.0
	github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue v1.0.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/pkg/sftp v1.13.7
	github.com/stretchr/testify v1.10.0
	github.com/yeka/zip v0.0.0-20231116150916-03d6312748a9
	golang.org/x/crypto v0.32.0
	golang.org/x/text v0.21.0

)

require (
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/internal v1.1.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
