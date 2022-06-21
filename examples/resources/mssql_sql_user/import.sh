# import using <db_id>/<user_id> - can be retrieved using `SELECT CONCAT(DB_ID(), '/', DATABASE_PRINCIPAL_ID('<username>'))`
terraform import mssql_sql_user.example '7/5'
