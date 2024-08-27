# resource for Azure Functions for SFTP
resource "azurerm_linux_function_app" "polling_trigger_function_app" {
  name                       = "polling-function-${var.environment}"
  location                   = data.azurerm_resource_group.group.location
  resource_group_name        = data.azurerm_resource_group.group.name
  service_plan_id            = azurerm_service_plan.plan.id
  storage_account_name       = azurerm_storage_account.storage.name
  storage_account_access_key = azurerm_storage_account.storage.primary_access_key
  https_only                 = true

  app_settings = {
    AZURE_STORAGE_CONNECTION_STRING = azurerm_storage_account.storage.primary_connection_string
    POLLING_TRIGGER_QUEUE_NAME      = azurerm_storage_queue.polling_trigger_queue.name
    CA_DPH_POLLING_CRON             = var.cron

    # Makes the Github Action run significantly faster by not copying the node_modules
    WEBSITE_RUN_FROM_PACKAGE = 1
  }

  site_config {
    #The below value should be kept at 1 so we don't duplicate actions and lock out the external sftp client
    app_scale_limit = 1

    # If `always_on` is not set to true, timers may only fire when an action (like a deploy
    # or looking at the app in the Azure Portal) causes the timers to sync
    always_on = true

    app_service_logs {
      retention_period_days = 60
    }

    application_stack {
      node_version = "20"
    }
  }

  lifecycle {
    ignore_changes = [
      # Ignore changes to tags because the CDC sets these automagically
      tags,
    ]
  }
}
