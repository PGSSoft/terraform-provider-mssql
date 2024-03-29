locals {
  public_ip = data.http.publicip_address.response_body
}

data "azurerm_client_config" "current" {}

data "azurerm_resource_group" "tests" {
  name = "terraform-mssql-tests"
}

data "http" "publicip_address" {
  url = "https://api.ipify.org/"
}