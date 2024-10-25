locals {
  environment_to_rs_environment_prefix_mapping = {
    dev = "staging"
    stg = "staging"
    prd = ""
  }
  selected_rs_environment_prefix = lookup(local.environment_to_rs_environment_prefix_mapping, var.environment, "staging")
  rs_domain_prefix               = "${local.selected_rs_environment_prefix}${length(local.selected_rs_environment_prefix) == 0 ? "" : "."}"
  higher_environment_level       = var.environment == "stg" || var.environment == "prd"
  cdc_domain_environment         = var.environment == "dev" || var.environment == "stg" || var.environment == "prd"
  non_pr_environment             = length(regexall("^pr\\d+", var.environment)) == 0
  cheezburger                    = "cheezburger" // setting something in TF to trigger a PR build
}

data "azurerm_resource_group" "group" {
  name = "csels-rsti-${var.environment}-moderate-rg"
}

data "azurerm_client_config" "current" {}
