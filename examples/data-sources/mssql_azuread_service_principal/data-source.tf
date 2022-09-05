data "mssql_database" "example" {
  name = "example"
}

data "mssql_azuread_service_principal" "example" {
  name        = "example"
  database_id = data.mssql_database.example.id
}

output "app_client_id" {
  value = data.mssql_azuread_service_principal.example.client_id
}
