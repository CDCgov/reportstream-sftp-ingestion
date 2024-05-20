resource "azurerm_storage_account" "storage" {
  name                            = "cdcrssftp${var.environment}"
  resource_group_name             = data.azurerm_resource_group.group.name
  location                        = data.azurerm_resource_group.group.location
  account_tier                    = "Standard"
  account_replication_type        = "GRS"
  account_kind                    = "StorageV2"
  allow_nested_items_to_be_public = false
}

resource "azurerm_storage_container" "sftp_container" {
  name                  = "sftp"
  storage_account_name  = azurerm_storage_account.storage.name
  container_access_type = "private"
}
