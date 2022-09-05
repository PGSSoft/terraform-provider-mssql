data "mssql_database" "example" {
  name = "example"
}

data "azuread_service_principal" "example" {
  display_name = "test-application"
}

resource "mssql_azuread_service_principal" "example" {
  name        = "example"
  database_id = data.mssql_database.example.id
  client_id   = data.azuread_service_principal.example.application_id
}

output "user_id" {
  value = mssql_azuread_service_principal.example.id
}
