package utils

import "testing"

func Test_populateConfig(t *testing.T) {
	jsonInput := []byte(`{
	"displayName": "Displayed Name",
	"isActive": "false"
}`)
	test := partnerConfig{}

	test.populatePartnerConfig(jsonInput)

}
