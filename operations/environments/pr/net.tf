resource "azurerm_virtual_network" "vnet" { //create the PR Vnet because it has a dynamic name that cannot be always pre-created
  name                = "csels-rsti-pr${var.pr_number}-moderate-sftp-app-vnet"
  location            = azurerm_resource_group.group.location
  resource_group_name = azurerm_resource_group.group.name

  address_space = ["10.0.0.0/25"]
}

resource "azurerm_network_security_group" "app_security_group" {
  name                = "sftp-app-security-group"
  location            = azurerm_resource_group.group.location
  resource_group_name = azurerm_resource_group.group.name
  lifecycle {
    ignore_changes = [
      # Ignore changes to tags because the CDC sets these automagically
      tags,
    ]
  }
}

resource "azurerm_network_security_rule" "App_Splunk_UF_omhsinf" {
  name                        = "Splunk_UF_omhsinf"
  priority                    = 103
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "9997-9998"
  source_address_prefixes     = ["10.65.8.211/32", "10.65.8.212/32", "10.65.7.212/32", "10.65.7.211/32", "10.65.8.210/32", "10.65.7.210/32"]
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.group.name
  network_security_group_name = azurerm_network_security_group.app_security_group.name
}

resource "azurerm_network_security_rule" "App_Splunk_Indexer_Discovery_omhsinf" {
  name                        = "Splunk_Indexer_Discovery_omhsinf"
  priority                    = 104
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "8089"
  source_address_prefix       = "10.11.7.22/32"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.group.name
  network_security_group_name = azurerm_network_security_group.app_security_group.name
}

resource "azurerm_network_security_rule" "App_Safe_Encase_Monitoring_omhsinf" {
  name                        = "Safe_Encase_Monitoring_omhsinf"
  priority                    = 105
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "34445"
  source_address_prefix       = "10.11.6.145/32"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.group.name
  network_security_group_name = azurerm_network_security_group.app_security_group.name
}

resource "azurerm_network_security_rule" "App_ForeScout_Manager_omhsinf" {
  name                        = "ForeScout_Manager_omhsinf"
  priority                    = 106
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_ranges     = ["556", "443", "10003-10006"]
  source_address_prefixes     = ["10.64.8.184", "10.64.8.180/32"]
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.group.name
  network_security_group_name = azurerm_network_security_group.app_security_group.name
}

resource "azurerm_network_security_rule" "App_BigFix_omhsinf" {
  name                        = "BigFix_omhsinf"
  priority                    = 107
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "52314"
  source_address_prefix       = "10.11.4.84/32"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.group.name
  network_security_group_name = azurerm_network_security_group.app_security_group.name
}

resource "azurerm_network_security_rule" "App_Allow_All_Out_omhsinf" {
  name                        = "Allow_All_Out_omhsinf"
  priority                    = 109
  direction                   = "Outbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  resource_group_name         = azurerm_resource_group.group.name
  network_security_group_name = azurerm_network_security_group.app_security_group.name
}
