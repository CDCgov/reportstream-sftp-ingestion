resource "azurerm_storage_account" "storage" {
  name                              = "cdcrssftp${var.environment}"
  resource_group_name               = data.azurerm_resource_group.group.name
  location                          = data.azurerm_resource_group.group.location
  account_tier                      = "Standard"
  account_replication_type          = "GRS"
  account_kind                      = "StorageV2"
  allow_nested_items_to_be_public   = false
  is_hns_enabled                    = true
  sftp_enabled                      = true
  min_tls_version                   = "TLS1_2"
  infrastructure_encryption_enabled = true

  lifecycle {
    ignore_changes = [
      customer_managed_key,
      tags, # Ignore changes to tags because the CDC sets these automagically
    ]
  }

  identity {
    type = "SystemAssigned"
  }
}


resource "azurerm_storage_account_customer_managed_key" "storage_storage_account_customer_key" {
  storage_account_id = azurerm_storage_account.storage.id
  key_vault_id       = azurerm_key_vault.key_storage.id
  key_name           = azurerm_key_vault_key.customer_managed_key.name

  depends_on = [
    azurerm_key_vault_access_policy.allow_github_deployer,
#     azurerm_key_vault_access_policy.allow_sftp_storage_account_wrapping
  ] //wait for the permission that allows our deployer to write the secret
}


resource "azurerm_storage_container" "sftp_container" {
  name                  = "sftp"
  storage_account_name  = azurerm_storage_account.storage.name
  container_access_type = "private"
}

resource "azurerm_storage_container" "sftp_container_dead_letter" {
  name                  = "sftp-dead-letter"
  storage_account_name  = azurerm_storage_account.storage.name
  container_access_type = "private"
}

resource "azurerm_role_assignment" "allow_app_read_write" {
  scope                = azurerm_storage_account.storage.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azurerm_linux_web_app.sftp.identity.0.principal_id
}

locals {
  # The below value needs to match the RS retention policy value so that we fall within their ATO.
  # Do not change it unless directed to do so by someone from the CDC security team
  retention_days = 60
}


resource "azurerm_storage_management_policy" "retention_policy" {
  storage_account_id = azurerm_storage_account.storage.id

  rule {
    name    = "pii_retention"
    enabled = true

    filters {
      blob_types = ["blockBlob", "appendBlob"]
    }

    actions {
      base_blob {
        delete_after_days_since_creation_greater_than = local.retention_days
      }
      snapshot {
        delete_after_days_since_creation_greater_than = local.retention_days
      }
      version {
        delete_after_days_since_creation = local.retention_days
      }
    }
  }
}
