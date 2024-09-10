# Create the container registry
resource "azurerm_container_registry" "registry" {
  name                = "cdcrssftp${var.environment}containerregistry"
  resource_group_name = data.azurerm_resource_group.group.name
  location            = data.azurerm_resource_group.group.location
  sku                 = "Premium"

  identity {
    type = "UserAssigned"
    identity_ids = [
      azurerm_user_assigned_identity.key_vault_identity.id
    ]
  }

  encryption {
    key_vault_key_id   = azurerm_key_vault_key.customer_managed_key.id
    identity_client_id = azurerm_user_assigned_identity.key_vault_identity.client_id
  }

  lifecycle {
    ignore_changes = [
      # Ignore changes to tags because the CDC sets these automagically
      tags,
    ]
  }

  depends_on = [azurerm_key_vault_access_policy.allow_container_registry_wrapping] // Wait for keyvault access policy to be in place before creating
}

resource "azurerm_role_assignment" "allow_app_to_pull_from_registry" {
  principal_id         = azurerm_linux_web_app.sftp.identity.0.principal_id
  role_definition_name = "AcrPull"
  scope                = azurerm_container_registry.registry.id
}

resource "azurerm_role_assignment" "allow_app_slot_to_pull_from_registry" {
  principal_id         = azurerm_linux_web_app_slot.pre_live.identity.0.principal_id
  role_definition_name = "AcrPull"
  scope                = azurerm_container_registry.registry.id
}

resource "azurerm_user_assigned_identity" "key_vault_identity" {
  resource_group_name = data.azurerm_resource_group.group.name
  location            = data.azurerm_resource_group.group.location

  name = "sftp-key-vault-identity-${var.environment}"

  lifecycle {
    ignore_changes = [
      # Ignore changes to tags because the CDC sets these automagically
      tags,
    ]
  }
}

# Create the staging service plan
resource "azurerm_service_plan" "plan" {
  name                   = "cdc-rs-sftp-${var.environment}-service-plan"
  resource_group_name    = data.azurerm_resource_group.group.name
  location               = data.azurerm_resource_group.group.location
  os_type                = "Linux"
  sku_name               = local.higher_environment_level ? "P1v3" : "P0v3"
  zone_balancing_enabled = local.higher_environment_level

  lifecycle {
    ignore_changes = [
      # Ignore changes to tags because the CDC sets these automagically
      tags,
    ]
  }
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
    health_check_path                 = "/health"
    health_check_eviction_time_in_min = 5

    container_registry_use_managed_identity = true

    scm_use_main_ip_restriction = local.cdc_domain_environment ? true : null

    application_stack {
      docker_registry_url = "https://${azurerm_container_registry.registry.login_server}"
      docker_image_name   = "ignore_because_specified_later_in_deployment"
    }

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

  #   When adding new settings that are needed for the live app but shouldn't be used in the pre-live
  #   slot, add them to `sticky_settings` as well as `app_settings` for the main app resource.
  #   All queue-related settings should be `sticky` so that the pre-live slot does not send or consume messages.
  app_settings = {
    WEBSITES_PORT = 8080
    PORT          = 8080

    ENV                             = var.environment
    AZURE_STORAGE_CONNECTION_STRING = azurerm_storage_account.storage.primary_blob_connection_string
    REPORT_STREAM_URL_PREFIX        = "https://${local.rs_domain_prefix}prime.cdc.gov"
    FLEXION_PRIVATE_KEY_NAME        = azurerm_key_vault_secret.mock_public_health_lab_private_key.name
    AZURE_KEY_VAULT_URI             = azurerm_key_vault.key_storage.vault_uri
    FLEXION_CLIENT_NAME             = "flexion.simulated-lab"
    QUEUE_MAX_DELIVERY_ATTEMPTS     = azurerm_eventgrid_system_topic_event_subscription.topic_sub.retry_policy.0.max_delivery_attempts # making the Azure container <-> queue retry count be in sync with the queue <-> application retry count..
  }

  sticky_settings {
    app_setting_names = ["AZURE_STORAGE_CONNECTION_STRING", "REPORT_STREAM_URL_PREFIX", "FLEXION_PRIVATE_KEY_NAME",
    "AZURE_KEY_VAULT_URI", "FLEXION_CLIENT_NAME", "QUEUE_MAX_DELIVERY_ATTEMPTS"]
  }

  identity {
    type = "SystemAssigned"
  }

  lifecycle {
    ignore_changes = [
      site_config[0].application_stack[0].docker_image_name,
      # Ignore changes to tags because the CDC sets these automagically
      tags,
    ]
  }
}

resource "azurerm_linux_web_app_slot" "pre_live" {
  name           = "pre-live"
  app_service_id = azurerm_linux_web_app.sftp.id

  https_only = true

  virtual_network_subnet_id = local.cdc_domain_environment ? azurerm_subnet.app.id : null

  site_config {
    health_check_path                 = "/health"
    health_check_eviction_time_in_min = 5

    scm_use_main_ip_restriction = local.cdc_domain_environment ? true : null

    container_registry_use_managed_identity = true

    application_stack {
      docker_registry_url = "https://${azurerm_container_registry.registry.login_server}"
      docker_image_name   = "ignore_because_specified_later_in_deployment"
    }

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
    WEBSITES_PORT = 8080
    PORT          = 8080

    ENV = var.environment
  }

  identity {
    type = "SystemAssigned"
  }

  lifecycle {
    ignore_changes = [
      site_config[0].application_stack[0].docker_image_name,
      # Ignore changes to tags because the CDC sets these automagically
      tags,
    ]
  }
}

resource "azurerm_monitor_autoscale_setting" "sftp_autoscale" {
  name                = "sftp_autoscale"
  resource_group_name = data.azurerm_resource_group.group.name
  location            = data.azurerm_resource_group.group.location
  target_resource_id  = azurerm_service_plan.plan.id

  lifecycle {
    ignore_changes = [
      # Ignore changes to tags because the CDC sets these automagically
      tags,
    ]
  }

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
