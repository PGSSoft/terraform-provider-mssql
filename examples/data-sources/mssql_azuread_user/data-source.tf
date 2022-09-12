data "mssql_database" "example" {
  name = "example"
}

data "mssql_azuread_user" "example" {
  name        = "example"
  database_id = data.mssql_database.example.id
}

output "user_object_id" {
  value = data.mssql_azuread_user.example.user_object_id
}
