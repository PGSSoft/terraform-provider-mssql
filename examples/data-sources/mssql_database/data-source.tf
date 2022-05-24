data "mssql_database" "example" {
  name = "example"
}

output "db_id" {
  value = data.mssql_database.example.id
}

output "db_collation" {
  value = data.mssql_database.example.collation
}