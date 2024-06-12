resource "azurerm_storage_account" "storage" {
  name                            = "cdcrssftp${var.environment}"
  resource_group_name             = data.azurerm_resource_group.group.name
  location                        = data.azurerm_resource_group.group.location
  account_tier                    = "Standard"
  account_replication_type        = "GRS"
  account_kind                    = "StorageV2"
  allow_nested_items_to_be_public = false
  is_hns_enabled                  = true
  sftp_enabled                    = true
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
  scope                = azurerm_storage_container.sftp_container.resource_manager_id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azurerm_linux_web_app.sftp.identity.0.principal_id
}

# The below is copied from ReportStream

# Not sure if we need any of the manual stuff referenced in this next comment

# Point-in-time restore, soft delete, versioning, and change feed were
# enabled in the portal as terraform does not currently support this.
# At some point, this should be moved into an azurerm_template_deployment
# resource.
# These settings can be configured under the "Data protection" blade
# for Blob service

variable "is_temp_env" {
  default = "false"
}

resource "azurerm_storage_management_policy" "retention_policy" {
  storage_account_id = azurerm_storage_account.storage.id

  rule {
    name    = "piiretention"
    enabled = true

    # TODO - update filters to be...?
    filters {
      prefix_match = ["reports/"]
      blob_types   = ["blockBlob", "appendBlob"]
    }

    actions {
      dynamic "base_blob" {
#         TODO - what envs should be temp? PR? Other ones?
        for_each = var.is_temp_env == false ? ["enabled"] : []

#         TODO - RS has 60 days for prod and staging and 30 days for demo and test
        content {
          delete_after_days_since_modification_greater_than = 60
        }
      }
      snapshot {
        delete_after_days_since_creation_greater_than = 60
      }
    }
  }
}
