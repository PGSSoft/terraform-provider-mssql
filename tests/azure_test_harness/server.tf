resource "azurerm_mssql_server" "this" {
  name                = local.name
  resource_group_name = data.azurerm_resource_group.tests.name
  location            = data.azurerm_resource_group.tests.location
  version             = "12.0"

  azuread_administrator {
    azuread_authentication_only = true
    login_username              = lookup(data.environment_variables.arm.items, "ARM_CLIENT_ID", data.azurerm_client_config.current.object_id)
    object_id                   = lookup(data.environment_variables.arm.items, "ARM_CLIENT_ID", data.azurerm_client_config.current.object_id)
  }
}

resource "azurerm_mssql_firewall_rule" "self" {
  name             = "test-runner"
  server_id        = azurerm_mssql_server.this.id
  start_ip_address = data.publicip_address.default.ip
  end_ip_address   = data.publicip_address.default.ip
}

resource "azurerm_mssql_elasticpool" "this" {
  name                = azurerm_mssql_server.this.name
  resource_group_name = azurerm_mssql_server.this.resource_group_name
  location            = azurerm_mssql_server.this.location
  server_name         = azurerm_mssql_server.this.name
  max_size_gb         = 5

  per_database_settings {
    max_capacity = 4
    min_capacity = 0.25
  }

  sku {
    capacity = 4
    name     = "GP_Gen5"
    tier     = "GeneralPurpose"
    family   = "Gen5"
  }
}