resource "azurerm_storage_queue" "message_queue" {
  name                 = "blob-message-queue"
  storage_account_name = azurerm_storage_account.storage.name
}
