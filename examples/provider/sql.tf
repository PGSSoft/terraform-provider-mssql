provider "mssql" {
  hostname = "localhost"
  port     = 1433

  sql_auth = {
    username = "sa"
    password = "sa_password"
  }
}
