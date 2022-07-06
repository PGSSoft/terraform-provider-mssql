data "mssql_database" "example" {
  name = "example"
}

data "mssql_sql_user" "owner" {
  name        = "example_user"
  database_id = data.mssql_database.example.id
}

data "mssql_sql_user" "member" {
  name        = "member_user"
  database_id = data.mssql_database.example.id
}

resource "mssql_database_role" "example" {
  name        = "example"
  database_id = data.mssql_database.example.id
  owner_id    = data.mssql_sql_user.owner.id
}

resource "mssql_datbase_role_member" "example" {
  role_id   = mssql_database_role.example.id
  member_id = data.mssql_sql_user.member.id
}