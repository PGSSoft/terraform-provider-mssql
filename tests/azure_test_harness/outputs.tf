resource "local_file" "env" {
  filename = "${path.module}/../../local.env"
  content = templatefile("${path.module}/env.tpl", {
    envs = {
      TF_ACC                     = 1
      TF_MSSQL_HOST              = azurerm_mssql_server.this.fully_qualified_domain_name
      TF_MSSQL_ELASTIC_POOL_NAME = azurerm_mssql_elasticpool.this.name
      TF_MSSQL_EDITION           = "azure"
      TF_MSSQL_MSI_NAME          = azurerm_user_assigned_identity.this.name
      TF_MSSQL_MSI_CLIENT_ID     = azurerm_user_assigned_identity.this.client_id
      TF_MSSQL_MSI_OBJECT_ID     = azurerm_user_assigned_identity.this.principal_id
    }
  })
}