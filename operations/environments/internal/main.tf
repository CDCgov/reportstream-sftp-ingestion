terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.116.0"
    }
  }

  # Use a remote Terraform state in Azure Storage
  backend "azurerm" {
    name                 = "test"
    resource_group_name  = "cdcti-terraform"
    storage_account_name = "cdctiterraform"
    container_name       = "tfstate"
    key                  = "rs-sftp-internal.terraform.tfstate"
  }
}

# Configure the Microsoft Azure Provider
provider "azurerm" {
  features {
    key_vault {
      purge_soft_deleted_secrets_on_destroy = false
    }
  }
}

module "template" {
  source = "../../template/"

  environment = "internal"
  deployer_id = "d59c2c86-de5e-41b7-a752-0869a73f5a60" //github app registration in Flexion Azure Entra
  cron        = "0 30 9 * * Wed"                       // Every Wednesday at 9:30 AM
}
