# import using <db_id>/<role_id> - can be retrieved using `SELECT CONCAT(DB_ID(), '/', DATABASE_PRINCIPAL_ID('<role_name>'))`
terraform import mssql_database_role.example '7/5'
