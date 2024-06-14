resource "azurerm_eventgrid_system_topic" "topic" {

  location               = "eastus"
  name                   = "blob-topic"
  resource_group_name    = data.azurerm_resource_group.group.name
  source_arm_resource_id = azurerm_storage_account.storage.id
  topic_type             = "Microsoft.Storage.StorageAccounts"

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_eventgrid_system_topic_event_subscription" "topic_sub" {

  name                = "topic-event-subscription"
  resource_group_name = data.azurerm_resource_group.group.name
  system_topic        = azurerm_eventgrid_system_topic.topic.name

  storage_queue_endpoint {
    queue_name                            = azurerm_storage_queue.message_queue.name
    storage_account_id                    = azurerm_storage_account.storage.id
    queue_message_time_to_live_in_seconds = 604800 # 7 days in seconds
  }

  included_event_types = ["Microsoft.Storage.BlobCreated"]

  advanced_filter {
    string_contains {
      key    = "subject"
      values = ["import"]
    }
  }

  retry_policy {
    event_time_to_live    = 1440 # 1 day in minutes
    max_delivery_attempts = 10
  }

  dead_letter_identity {
    type = "SystemAssigned"
  }

  storage_blob_dead_letter_destination {
    storage_account_id          = azurerm_storage_account.storage.id
    storage_blob_container_name = azurerm_storage_container.sftp_container_dead_letter.name
  }

  depends_on = [azurerm_role_assignment.allow_event_read_write]
}

resource "azurerm_role_assignment" "allow_event_read_write" {
  scope                = azurerm_storage_account.storage.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azurerm_eventgrid_system_topic.topic.identity.0.principal_id
}
