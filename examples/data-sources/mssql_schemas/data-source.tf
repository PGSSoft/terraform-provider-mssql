data "mssql_database" "example" {
  name = "example"
}

data "mssql_schemas" "all" {
  database_id = data.mssql_database.example.id
}

output "all_schema_names" {
  value = data.mssql_schemas.all.schemas[*].name
}