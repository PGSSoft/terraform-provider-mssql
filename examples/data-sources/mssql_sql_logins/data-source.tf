data "mssql_sql_logins" "example" {}

output "databases" {
  value = data.mssql_sql_logins.example.logins
}