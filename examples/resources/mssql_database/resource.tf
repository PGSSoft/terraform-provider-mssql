resource "mssql_database" "example" {
  name      = "example"
  collation = "SQL_Latin1_General_CP1_CS_AS"
}