resource "azurerm_user_assigned_identity" "this" {
  name                = local.name
  resource_group_name = data.azurerm_resource_group.tests.name
  location            = data.azurerm_resource_group.tests.location
}

data "azurerm_user_assigned_identity" "server" {
  name                = "terraform-mssql-tests"
  resource_group_name = data.azurerm_resource_group.tests.name
}