# resource for Azure Functions for SFTP
resource "azurerm_linux_function_app" "polling_trigger_function_app" {
  name                       = "polling-function-${var.environment}"
  location                   = data.azurerm_resource_group.group.location
  resource_group_name        = data.azurerm_resource_group.group.name
  service_plan_id            = azurerm_service_plan.plan.id
  storage_account_name       = azurerm_storage_account.storage.name
  storage_account_access_key = azurerm_storage_account.storage.primary_access_key

  app_settings = {
    AZURE_STORAGE_CONNECTION_STRING = azurerm_storage_account.storage.primary_blob_connection_string
    POLLING_TRIGGER_QUEUE_NAME      = azurerm_storage_queue.polling_trigger_queue.name
  }

  site_config {
    app_scale_limit = 1
    application_insights_connection_string = azurerm_application_insights.function_app_insights.connection_string
    application_insights_key = azurerm_application_insights.function_app_insights.instrumentation_key

    # TODO - verify this is good advice
    always_on = true

    app_service_logs {
      retention_period_days = 60
    }

    application_stack {
      node_version = "20"
    }
  }
}

resource "azurerm_application_insights" "function_app_insights" {
  name                = "functionapp-insights-${var.environment}"
  location            = data.azurerm_resource_group.group.location
  resource_group_name = data.azurerm_resource_group.group.name
  workspace_id        = azurerm_log_analytics_workspace.logs_workspace.id
  application_type    = "Node.JS"
}
