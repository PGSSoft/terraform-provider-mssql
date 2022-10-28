# import using <db_id>/<principal_id>/<permission> - can be retrieved using `SELECT CONCAT(DB_ID(), '/', DATABASE_PRINCIPAL_ID('<principal_name>'), '/DELETE')`
terraform import mssql_database_permission.example '7/5/DELETE'
