resource "azurerm_storage_queue" "message_import_queue" {
  name                 = "message-import-queue"
  storage_account_name = azurerm_storage_account.storage.name
}

resource "azurerm_storage_queue" "message_import_dead_letter_queue" {
  name                 = "message-import-dead-letter-queue"
  storage_account_name = azurerm_storage_account.storage.name
}

resource "azurerm_storage_queue" "polling_trigger_queue" {
  name                 = "polling-trigger-queue"
  storage_account_name = azurerm_storage_account.storage.name
}

resource "azurerm_storage_queue" "polling_trigger_dead_letter_queue" {
  name                 = "polling-trigger-dead-letter-queue"
  storage_account_name = azurerm_storage_account.storage.name
}
