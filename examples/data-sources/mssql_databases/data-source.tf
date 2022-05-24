data "mssql_databases" "example" {}

output "databases" {
  value = data.mssql_databases.example.databases
}