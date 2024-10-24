data "azurerm_monitor_action_group" "notify_slack_email" {
  count               = local.non_pr_environment ? 1 : 0
  resource_group_name = data.azurerm_resource_group.group.name
  name                = "cdcti${var.environment}-actiongroup"
}

resource "azurerm_monitor_metric_alert" "azure_4XX_alert" {
  count               = local.non_pr_environment ? 1 : 0
  name                = "cdc-rs-sftp-${var.environment}-azure-http-4XX-alert"
  resource_group_name = data.azurerm_resource_group.group.name
  scopes              = [azurerm_linux_web_app.sftp.id]
  description         = "Action will be triggered when Http Status Code 4XX is greater than or equal to 3"
  frequency           = "PT1M" // Checks every 1 minute
  window_size         = "PT1H" // Every Check looks back 1 hour for 4xx errors

  criteria {
    metric_namespace = "Microsoft.Web/sites"
    metric_name      = "Http4xx"
    aggregation      = "Total"
    operator         = "GreaterThanOrEqual"
    threshold        = 3
  }

  action {
    action_group_id = data.azurerm_monitor_action_group.notify_slack_email[count.index].id
  }

  lifecycle {
    # Ignore changes to tags because the CDC sets these automagically
    ignore_changes = [
      tags["business_steward"],
      tags["center"],
      tags["environment"],
      tags["escid"],
      tags["funding_source"],
      tags["pii_data"],
      tags["security_compliance"],
      tags["security_steward"],
      tags["support_group"],
      tags["system"],
      tags["technical_steward"],
      tags["zone"]
    ]
  }
}

resource "azurerm_monitor_metric_alert" "azure_5XX_alert" {
  count               = local.non_pr_environment ? 1 : 0
  name                = "cdc-rs-sftp-${var.environment}-azure-http-5XX-alert"
  resource_group_name = data.azurerm_resource_group.group.name
  scopes              = [azurerm_linux_web_app.sftp.id]
  description         = "Action will be triggered when Http Status Code 5XX is greater than or equal to 1"
  frequency           = "PT1M" // Checks every 1 minute
  window_size         = "PT5M" // Every Check looks back 5 min for 5xx errors

  criteria {
    metric_namespace = "Microsoft.Web/sites"
    metric_name      = "Http5xx"
    aggregation      = "Total"
    operator         = "GreaterThanOrEqual"
    threshold        = 1
  }

  action {
    action_group_id = data.azurerm_monitor_action_group.notify_slack_email[count.index].id
  }

  lifecycle {
    # Ignore changes to tags because the CDC sets these automagically
    ignore_changes = [
      tags["business_steward"],
      tags["center"],
      tags["environment"],
      tags["escid"],
      tags["funding_source"],
      tags["pii_data"],
      tags["security_compliance"],
      tags["security_steward"],
      tags["support_group"],
      tags["system"],
      tags["technical_steward"],
      tags["zone"]
    ]
  }
}