resource "azurerm_eventgrid_system_topic" "topic" {

  location               = "eastus"
  name                   = "blob-topic"
  resource_group_name    = data.azurerm_resource_group.group.name
  source_arm_resource_id = azurerm_storage_account.storage.id
  topic_type             = "Microsoft.Storage.StorageAccounts"
}

resource "azurerm_eventgrid_system_topic_event_subscription" "topic_sub" {

  name                = "topic-event-subscription"
  resource_group_name = data.azurerm_resource_group.group.name
  system_topic        = azurerm_eventgrid_system_topic.topic.name

  storage_queue_endpoint {
    queue_name         = azurerm_storage_queue.message_queue.name
    storage_account_id = azurerm_storage_account.storage.id
    queue_message_time_to_live_in_seconds = 604800 # in seconds
  }

  included_event_types = ["Microsoft.Storage.BlobCreated"]

  advanced_filter {
    string_contains {
      key = "subject"
      values = ["import"]
    }
  }

  retry_policy {
    event_time_to_live    = 1440 #in minutes
    max_delivery_attempts = 10
  }

  dead_letter_identity {
    type = "SystemAssigned"
  }

  storage_blob_dead_letter_destination {
    storage_account_id          = azurerm_storage_account.storage.id
    storage_blob_container_name = azurerm_storage_queue.message_dead_letter_queue.name
  }
}
