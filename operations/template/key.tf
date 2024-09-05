resource "azurerm_key_vault" "key_storage" {
  name = "rs-sftp-vault-${var.environment}"

  resource_group_name = data.azurerm_resource_group.group.name
  location            = data.azurerm_resource_group.group.location

  sku_name  = "standard"
  tenant_id = data.azurerm_client_config.current.tenant_id

  purge_protection_enabled = true

  lifecycle {
    ignore_changes = [
      # Ignore changes to tags because the CDC sets these automagically
      tags,
    ]
  }
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

  key_permissions = [
    "Create",
    "Delete",
    "Get",
    "Purge",
    "Recover",
    "Update",
    "GetRotationPolicy",
    "SetRotationPolicy",
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

resource "azurerm_key_vault_access_policy" "allow_sftp_storage_account_wrapping" {
  key_vault_id = azurerm_key_vault.key_storage.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = azurerm_storage_account.storage.identity.0.principal_id

  key_permissions = [
    "Get",
    "UnwrapKey",
    "WrapKey",
  ]
}

resource "azurerm_key_vault_access_policy" "allow_container_registry_wrapping" {
  key_vault_id = azurerm_key_vault.key_storage.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = azurerm_user_assigned_identity.key_vault_identity.principal_id

  key_permissions = [
    "Get",
    "UnwrapKey",
    "WrapKey",
  ]
}


resource "azurerm_key_vault_secret" "mock_public_health_lab_private_key" {
  name  = "mock-public-health-lab-private-key-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}

resource "azurerm_key_vault_secret" "ca_phl_private_key" {
  name  = "ca-phl-private-key-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}

resource "azurerm_key_vault_secret" "ca_phl_zip_password" {
  name  = "ca-phl-zip-password-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}

resource "azurerm_key_vault_secret" "ca_phl_sftp_starting_directory" {
  name  = "ca-phl-sftp-starting-directory-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}

resource "azurerm_key_vault_secret" "ca_phl_sftp_user" {
  name  = "ca-phl-sftp-user-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}

resource "azurerm_key_vault_secret" "ca_phl_sftp_key" {
  name  = "ca-phl-sftp-key-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}

resource "azurerm_key_vault_secret" "ca_phl_sftp_server_address" {
  name  = "ca-phl-sftp-server-address-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}

resource "azurerm_key_vault_secret" "ca_phl_sftp_server_public_key" {
  name  = "ca-phl-sftp-server-public-key-${var.environment}"
  value = "dogcow"

  key_vault_id = azurerm_key_vault.key_storage.id

  lifecycle {
    ignore_changes = [value]
  }
  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}

resource "azurerm_key_vault_key" "customer_managed_key" {
  name         = "customer-managed-key-${var.environment}"
  key_vault_id = azurerm_key_vault.key_storage.id

  key_type = "RSA"
  key_size = 4096

  key_opts = [
    "decrypt",
    "encrypt",
    "sign",
    "unwrapKey",
    "verify",
    "wrapKey"
  ]

  depends_on = [azurerm_key_vault_access_policy.allow_github_deployer] //wait for the permission that allows our deployer to write the secret
}