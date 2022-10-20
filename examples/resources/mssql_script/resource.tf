data "mssql_database" "test" {
  name = "test"
}

resource "mssql_script" "cdc" {
  database_id = data.mssql_database.test.id

  read_script   = "SELECT COUNT(*) AS [is_enabled] FROM sys.change_tracking_databases WHERE database_id=${data.mssql_database.test.id}"
  delete_script = "ALTER DATABASE [${data.mssql_database.test.name}] SET CHANGE_TRACKING = OFF"

  update_script = <<SQL
IF (SELECT COUNT(*) FROM sys.change_tracking_databases WHERE database_id=${data.mssql_database.test.id}) = 0
  ALTER DATABASE [${data.mssql_database.test.name}] SET CHANGE_TRACKING = ON
SQL

  state = {
    is_enabled = "1"
  }
}