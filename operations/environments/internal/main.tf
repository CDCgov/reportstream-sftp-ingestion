terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.16.0"
    }
  }

  # Use a remote Terraform state in Azure Storage
  backend "azurerm" {
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
      purge_soft_deleted_keys_on_destroy    = false
    }
  }
}

module "template" {
  source = "../../template/"

  environment = "internal"
  deployer_id = "d59c2c86-de5e-41b7-a752-0869a73f5a60" //github app registration in Flexion Azure Entra
  cron        = "* * * 30 Feb *"                       //run every second of February 30th, which never happens and is the equivalent of never running.  If you want to run this, manually trigger the function in Azure.
}
