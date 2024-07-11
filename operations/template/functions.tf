# resource for Azure Functions for SFTP
resource "azurerm_linux_function_app" "polling_trigger_function_app" {
  name                       = "polling-function-${var.environment}"
  location                   = data.azurerm_resource_group.group.location
  resource_group_name        = data.azurerm_resource_group.group.name
  service_plan_id            = azurerm_service_plan.plan.id
  storage_account_name       = azurerm_storage_account.storage.name
  storage_account_access_key = azurerm_storage_account.storage.primary_access_key

  site_config {
    app_scale_limit = 1

    # TODO - verify this is good advice
    # always_on = true

    app_service_logs {
      retention_period_days = 60
    }

    application_stack {
      node_version = "20"
    }
  }


}
