# import using <db_id>/<schema_id> - can be retrieved using `SELECT CONCAT(DB_ID(), '/', SCHEMA_ID('<schema_name>'))`
terraform import mssql_schema.example '7/5'
