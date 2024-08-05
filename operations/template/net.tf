data "azurerm_virtual_network" "app" {
  name                = "csels-rsti-${var.environment}-moderate-sftp-app-vnet"
  resource_group_name = data.azurerm_resource_group.group.name
}

data "azurerm_network_security_group" "app_security_group" {
  name                = "sftp-app-security-group"
  location            = data.azurerm_resource_group.group.location
  resource_group_name = data.azurerm_resource_group.group.name
}

locals {
  subnets_cidrs = cidrsubnets(data.azurerm_virtual_network.app.address_space[0], 2)
}

resource "azurerm_subnet" "app" {
  name                 = "sftp-app"
  resource_group_name  = data.azurerm_resource_group.group.name
  virtual_network_name = data.azurerm_virtual_network.app.name
  address_prefixes     = [local.subnets_cidrs[0]]

  service_endpoints = [
    "Microsoft.AzureActiveDirectory",
    "Microsoft.AzureCosmosDB",
    "Microsoft.ContainerRegistry",
    "Microsoft.EventHub",
    "Microsoft.KeyVault",
    "Microsoft.ServiceBus",
    "Microsoft.Sql",
    "Microsoft.Storage",
    "Microsoft.Web",
  ]

  delegation {
    name = "delegation"

    service_delegation {
      name    = "Microsoft.Web/serverFarms"
      actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_subnet_network_security_group_association" "app_security_group" {
  subnet_id                 = azurerm_subnet.app.id
  network_security_group_id = data.azurerm_network_security_group.app_security_group.id
}
