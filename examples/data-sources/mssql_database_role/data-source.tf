data "mssql_database" "master" {
  name = "master"
}

data "mssql_database_role" "example" {
  name        = "public"
  database_id = data.mssql_database.master.id
}

output "id" {
  value = data.mssql_database_role.example.id
}
