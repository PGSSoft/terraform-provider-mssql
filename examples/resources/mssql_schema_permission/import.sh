# import using <db_id>/<schema_id>/<principal_id>/<permission> - can be retrieved using `SELECT CONCAT(DB_ID(), '/', SCHEMA_ID('<schema_name>'), '/', DATABASE_PRINCIPAL_ID('<principal_name>'), '/DELETE')`
terraform import mssql_schema_permission.example '7/5/8/DELETE'
