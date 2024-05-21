resource "azurerm_key_vault" "key_storage" {
  name = "rs-sftp-key-vault-${var.environment}"

  resource_group_name = data.azurerm_resource_group.group.name
  location            = data.azurerm_resource_group.group.location

  sku_name  = "standard"
  tenant_id = data.azurerm_client_config.current.tenant_id

  purge_protection_enabled = false
}

resource "azurerm_key_vault_access_policy" "allow_github_deployer" {
  key_vault_id = azurerm_key_vault.key_storage.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = var.deployer_id

  secret_permissions = [
    "Set",
    "Get",
    "Delete",
    "Purge",
  ]
}

resource "azurerm_key_vault_access_policy" "allow_app_read" {
  key_vault_id = azurerm_key_vault.key_storage.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = azurerm_linux_web_app.sftp.identity.0.principal_id

  secret_permissions = [
    "List",
    "Get",
  ]
}

resource "azurerm_key_vault_secret" "report_stream_public_key" {
  name  = "organization-report-stream-public-key-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}

resource "azurerm_key_vault_secret" "trusted_intermediary_public_key" {
  name  = "organization-trusted-intermediary-public-key-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}

resource "azurerm_key_vault_secret" "trusted_intermediary_public_key_internal" {
  name  = "trusted-intermediary-public-key-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}

resource "azurerm_key_vault_secret" "trusted_intermediary_private_key" {
  name  = "trusted-intermediary-private-key-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}
