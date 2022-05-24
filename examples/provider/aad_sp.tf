provider "mssql" {
  hostname = "example.database.windows.net"
  port     = 1433

  azure_auth = {
    client_id     = "94e8d55d-cbbc-4e41-b21a-8923d83f9a85"
    client_secret = "client_secret"
    tenant_id     = "a352c914-bfd9-4b7e-8b1d-554a58353f22"
  }
}
