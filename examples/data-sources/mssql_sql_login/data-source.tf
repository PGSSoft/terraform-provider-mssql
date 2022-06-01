data "mssql_sql_login" "sa" {
  name = "sa"
}

output "id" {
  value = data.mssql_sql_login.sa.id
}

output "db_id" {
  value = data.mssql_sql_login.sa.default_database_id
}