resource "azurerm_log_analytics_workspace" "logs_workspace" {
  name = "rs-sftp-logs-${var.environment}"

  resource_group_name = data.azurerm_resource_group.group.name
  location            = data.azurerm_resource_group.group.location

  lifecycle {
    ignore_changes = [
      # Ignore changes to tags because the CDC sets these automagically
      tags,
    ]
  }
}

resource "azurerm_log_analytics_query_pack" "application_logs_pack" {
  name                = "RS SFTP Application Logs"
  resource_group_name = data.azurerm_resource_group.group.name
  location            = data.azurerm_resource_group.group.location

  lifecycle {
    ignore_changes = [
      # Ignore changes to tags because the CDC sets these automagically
      tags,
    ]
  }
}

resource "azurerm_log_analytics_query_pack_query" "structured_application_logs" {
  display_name = "RS SFTP's Raw Application Logs"
  description  = "View all RS SFTP's application logs in a structured format"

  query_pack_id = azurerm_log_analytics_query_pack.application_logs_pack.id
  categories    = ["applications"]

  body = "AppServiceConsoleLogs | project JsonResult = parse_json(ResultDescription) | evaluate bag_unpack(JsonResult) | project-reorder ['time'], level, msg"
}

resource "azurerm_monitor_diagnostic_setting" "app_to_logs" {
  name                       = "rs-sftp-app-to-logs-${var.environment}"
  target_resource_id         = azurerm_linux_web_app.sftp.id
  log_analytics_workspace_id = azurerm_log_analytics_workspace.logs_workspace.id

  log_analytics_destination_type = "Dedicated"

  enabled_log {
    category = "AppServiceConsoleLogs"
  }
  enabled_log {
    category = "AppServiceAppLogs"
  }
  enabled_log {
    category = "AppServiceHTTPLogs"
  }
  enabled_log {
    category = "AppServicePlatformLogs"
  }
}

resource "azurerm_monitor_diagnostic_setting" "prelive_slot_to_logs" {
  name                       = "rs-sftp-prelive-slot-to-logs-${var.environment}"
  target_resource_id         = azurerm_linux_web_app_slot.pre_live.id
  log_analytics_workspace_id = azurerm_log_analytics_workspace.logs_workspace.id

  log_analytics_destination_type = "Dedicated"

  enabled_log {
    category = "AppServiceConsoleLogs"
  }
  enabled_log {
    category = "AppServiceAppLogs"
  }
  enabled_log {
    category = "AppServiceHTTPLogs"
  }
  enabled_log {
    category = "AppServicePlatformLogs"
  }
}

resource "azurerm_monitor_diagnostic_setting" "functionapp_to_logs" {
  name                       = "rs-sftp-function-app-to-logs-${var.environment}"
  target_resource_id         = azurerm_linux_function_app.polling_trigger_function_app.id
  log_analytics_workspace_id = azurerm_log_analytics_workspace.logs_workspace.id

  log_analytics_destination_type = "Dedicated"

  enabled_log {
    category = "FunctionAppLogs"
  }
}
