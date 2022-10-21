data "mssql_database" "test" {
  name = "test"
}

data "mssql_query" "column" {
  database_id = data.mssql_database.test.id
  query       = "SELECT [column_id], [name] FROM sys.columns WHERE [object_id] = OBJECT_ID('test_table')"
}

output "column_names" {
  value = data.mssql_query.column.result[*].name
}

