data "mssql_sql_login" "member" {
  name = "member_login"
}

resource "mssql_server_role" "example" {
  name = "example"
}

resource "mssql_server_role_member" "example" {
  role_id   = mssql_server_role.example.id
  member_id = data.mssql_sql_login.member.id
}