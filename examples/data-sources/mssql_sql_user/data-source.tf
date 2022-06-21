data "mssql_database" "master" {
  name = "master"
}

data "mssql_sql_user" "example" {
  name        = "dbo"
  database_id = data.mssql_database.master.id
}

output "id" {
  value = data.mssql_sql_user.example.id
}
