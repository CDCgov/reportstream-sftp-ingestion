# Create the container registry
resource "azurerm_container_registry" "registry" {
  name                = "cdcrssftp${var.environment}containerregistry"
  resource_group_name = data.azurerm_resource_group.group.name
  location            = data.azurerm_resource_group.group.location
  sku                 = "Standard"
  admin_enabled       = true
}

# Create the staging service plan
resource "azurerm_service_plan" "plan" {
  name                   = "cdc-rs-sftp-${var.environment}-service-plan"
  resource_group_name    = data.azurerm_resource_group.group.name
  location               = data.azurerm_resource_group.group.location
  os_type                = "Linux"
  sku_name               = local.higher_environment_level ? "P1v3" : "P0v3"
  zone_balancing_enabled = local.higher_environment_level
}

# Create the staging App Service
resource "azurerm_linux_web_app" "sftp" {
  name                = "cdc-rs-sftp-${var.environment}"
  resource_group_name = data.azurerm_resource_group.group.name
  location            = azurerm_service_plan.plan.location
  service_plan_id     = azurerm_service_plan.plan.id

  https_only = true

  virtual_network_subnet_id = local.cdc_domain_environment ? azurerm_subnet.app.id : null

  site_config {
    scm_use_main_ip_restriction = local.cdc_domain_environment ? true : null

    dynamic "ip_restriction" {
      for_each = local.cdc_domain_environment ? [1] : []

      content {
        name       = "deny_all_ipv4"
        action     = "Deny"
        ip_address = "0.0.0.0/0"
        priority   = "200"
      }
    }

    dynamic "ip_restriction" {
      for_each = local.cdc_domain_environment ? [1] : []

      content {
        name       = "deny_all_ipv6"
        action     = "Deny"
        ip_address = "::/0"
        priority   = "201"
      }
    }
  }

  app_settings = {
    DOCKER_REGISTRY_SERVER_URL      = "https://${azurerm_container_registry.registry.login_server}"
    DOCKER_REGISTRY_SERVER_USERNAME = azurerm_container_registry.registry.admin_username
    DOCKER_REGISTRY_SERVER_PASSWORD = azurerm_container_registry.registry.admin_password
    ENV                             = var.environment
    AZURE_BLOB_CONNECTION_STRING    = azurerm_storage_account.storage.primary_blob_connection_string
    REPORT_STREAM_URL_PREFIX        = "https://${local.rs_domain_prefix}prime.cdc.gov"
    FLEXION_PRIVATE_KEY_NAME        = azurerm_key_vault_secret.mock_public_health_lab_private_key.name
    AZURE_KEY_VAULT_URI             = azurerm_key_vault.key_storage.vault_uri
    FLEXION_CLIENT_NAME             = "flexion.simulated-lab"
    AZURE_SDK_GO_LOGGING            = "all"
  }

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_monitor_autoscale_setting" "sftp_autoscale" {
  name                = "sftp_autoscale"
  resource_group_name = data.azurerm_resource_group.group.name
  location            = data.azurerm_resource_group.group.location
  target_resource_id  = azurerm_service_plan.plan.id


  profile {
    name = "defaultProfile"

    capacity {
      default = local.higher_environment_level ? 3 : 1
      minimum = local.higher_environment_level ? 3 : 1
      maximum = local.higher_environment_level ? 10 : 1
    }

    rule {
      metric_trigger {
        metric_name        = "CpuPercentage"
        metric_resource_id = azurerm_service_plan.plan.id
        time_grain         = "PT1M"
        statistic          = "Average"
        time_window        = "PT5M"
        time_aggregation   = "Average"
        operator           = "GreaterThan"
        threshold          = 75
        metric_namespace   = "microsoft.web/serverfarms"
      }

      scale_action {
        direction = "Increase"
        type      = "ChangeCount"
        value     = "1"
        cooldown  = "PT1M"
      }
    }

    rule {
      metric_trigger {
        metric_name        = "CpuPercentage"
        metric_resource_id = azurerm_service_plan.plan.id
        time_grain         = "PT1M"
        statistic          = "Average"
        time_window        = "PT5M"
        time_aggregation   = "Average"
        operator           = "LessThan"
        threshold          = 25
        metric_namespace   = "microsoft.web/serverfarms"
      }

      scale_action {
        direction = "Decrease"
        type      = "ChangeCount"
        value     = "1"
        cooldown  = "PT5M"
      }
    }
  }
}
