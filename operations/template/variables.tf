variable "environment" {
  type     = string
  nullable = false
}

variable "deployer_id" {
  type     = string
  nullable = false
}


variable "cron" {
  type     = string
  nullable = false
}

variable "alert_slack_email" {
  type      = string
  nullable  = false
  sensitive = true
}