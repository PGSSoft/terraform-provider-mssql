# import using <db_id>/<user_id> - can be retrieved using `SELECT CONCAT(DB_ID(), '/', principal_id) FROM sys.database_principals WHERE [name] = '<username>'`
terraform import mssql_azuread_service_principal.example '7/5'
