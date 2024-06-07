resource "azurerm_storage_queue" "message_queue" {
  name                 = "blob-message-queue"
  storage_account_name = azurerm_storage_account.storage.name
}

resource "azurerm_storage_queue" "message_dead_letter_queue" {
  name                 = "blob-message-dead-letter-queue"
  storage_account_name = azurerm_storage_account.storage.name
}
