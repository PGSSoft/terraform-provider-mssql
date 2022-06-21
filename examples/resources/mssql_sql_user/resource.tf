data "mssql_database" "example" {
  name = "example"
}

resource "mssql_sql_login" "example" {
  name                      = "example"
  password                  = "Str0ngPa$$word12"
  must_change_password      = true
  default_database_id       = data.mssql_database.example.id
  default_language          = "english"
  check_password_expiration = true
  check_password_policy     = true
}

resource "mssql_sql_user" "example" {
  name        = "example"
  database_id = data.mssql_database.example.id
  login_id    = mssql_sql_login.example.id
}

output "user_id" {
  value = mssql_sql_user.example.id
}