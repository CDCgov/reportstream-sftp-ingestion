data "azurerm_virtual_network" "app" {
  name                = "csels-rs-sftp-${var.environment}-moderate-app-vnet"
  resource_group_name = data.azurerm_resource_group.group.name
}

locals {
  subnets_cidrs = cidrsubnets(data.azurerm_virtual_network.app.address_space[0], 2, 2, 2, 3, 3)
}

#
# resource "azurerm_subnet" "app" {
#   name                 = "app"
#   resource_group_name  = data.azurerm_resource_group.group.name
#   virtual_network_name = data.azurerm_virtual_network.app.name
#   address_prefixes     = [local.subnets_cidrs[0]]
#
#   service_endpoints = [
#     "Microsoft.AzureActiveDirectory",
#     "Microsoft.AzureCosmosDB",
#     "Microsoft.ContainerRegistry",
#     "Microsoft.EventHub",
#     "Microsoft.KeyVault",
#     "Microsoft.ServiceBus",
#     "Microsoft.Sql",
#     "Microsoft.Storage",
#     "Microsoft.Web",
#   ]
#
#   delegation {
#     name = "delegation"
#
#     service_delegation {
#       name    = "Microsoft.Web/serverFarms"
#       actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
#     }
#   }
# }