data "mssql_database" "example" {
  name = "example"
}

data "mssql_schema" "by_name" {
  database_id = data.mssql_database.example.id
  name        = "dbo"
}