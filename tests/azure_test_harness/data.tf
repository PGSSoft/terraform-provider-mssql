data "azurerm_client_config" "current" {}

data "azurerm_resource_group" "tests" {
  name = "terraform-mssql-tests"
}

data "publicip_address" "default" {
  source_ip = "0.0.0.0"
}