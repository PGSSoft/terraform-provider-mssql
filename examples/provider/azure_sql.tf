terraform {
  required_providers {
    publicip = {
      source = "nxt-engineering/publicip"
    }

    mssql = {
      source = "PGSSoft/mssql"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "mssql" {
  hostname   = azurerm_mssql_server.this.fully_qualified_domain_name
  azure_auth = {}
}

data "azurerm_client_config" "current" {}

data "publicip_address" "default" {
  source_ip = "0.0.0.0"
}

resource "azurerm_resource_group" "azure_test" {
  name     = "terraform-mssql-test-local"
  location = "WestEurope"
}

resource "azurerm_mssql_server" "this" {
  name                = "terraform-mssql-test-local"
  resource_group_name = azurerm_resource_group.azure_test.name
  location            = azurerm_resource_group.azure_test.location
  version             = "12.0"

  azuread_administrator {
    login_username              = data.azurerm_client_config.current.object_id
    object_id                   = data.azurerm_client_config.current.object_id
    azuread_authentication_only = true
  }
}

resource "azurerm_mssql_firewall_rule" "caller" {
  name             = "caller"
  server_id        = azurerm_mssql_server.this.id
  start_ip_address = data.publicip_address.default.ip
  end_ip_address   = data.publicip_address.default.ip
}

resource "azurerm_mssql_database" "test" {
  name      = "test"
  server_id = azurerm_mssql_server.this.id
}

data "mssql_database" "test" {
  name = azurerm_mssql_database.test.name
}

data "mssql_database_role" "test_owner" {
  database_id = data.mssql_database.test.id
  name        = "db_owner"
}

resource "mssql_azuread_user" "caller" {
  name           = data.azurerm_client_config.current.object_id
  database_id    = data.mssql_database.test.id
  user_object_id = data.azurerm_client_config.current.object_id
}

resource "mssql_database_role_member" "caller_owner" {
  role_id   = data.mssql_database_role.test_owner.id
  member_id = mssql_azuread_user.caller.id
}