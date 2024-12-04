package utils

import "testing"

func Test_populateConfig(t *testing.T) {
	jsonInput := []byte(`{
	"displayName": "Displayed Name",
	"isActive": true
}`)
	test := partnerConfig{}

	test.populatePartnerConfig(jsonInput)

}
func Test_populateConfigEntry(t *testing.T) {
	jsonInput := []byte(`{
	"partnerId": "ca-phl",
	"partnerSettings": {
	"displayName": "Displayed Name",
	"isActive": true,
	"sftpConnectionType": "external",
	"hasZipPassword": true,
	"defaultEncoding": "ISO-8859-1",
	"cronExpression": "* * * * *",
	"containerName": "blah"
}}`)
	test := partnerConfig{}

	_, _ = test.populateConfigEntry(jsonInput)

}
