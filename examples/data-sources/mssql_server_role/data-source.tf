data "mssql_server_role" "by_name" {
  name = "example"
}

data "mssql_server_role" "by_id" {
  id = 8
}