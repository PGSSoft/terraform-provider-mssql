data "mssql_database" "example" {
  name = "example"
}

data "azuread_user" "example" {
  user_principal_name = "user@example.com"
}

resource "mssql_azuread_user" "example" {
  name           = "example"
  database_id    = data.mssql_database.example.id
  user_object_id = data.azuread_user.example.object_id
}

output "user_id" {
  value = mssql_azuread_user.example.id
}