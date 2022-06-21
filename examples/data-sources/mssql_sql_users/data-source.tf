data "mssql_database" "master" {
  name = "master"
}

data "mssql_sql_users" "example" {
  database_id = data.mssql_database.master.id
}

output "users" {
  value = data.mssql_sql_users.example.users
}
