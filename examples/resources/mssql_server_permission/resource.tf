data "mssql_sql_login" "example" {
  name = "example_login"
}

resource "mssql_server_permission" "connect_to_example" {
  principal_id      = data.mssql_sql_login.example.principal_id
  permission        = "CONNECT SQL"
  with_grant_option = true
}