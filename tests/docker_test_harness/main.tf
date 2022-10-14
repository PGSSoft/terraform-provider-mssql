resource "random_password" "mssql" {
  length  = 20
  special = false
}

resource "docker_image" "mssql" {
  name         = "mcr.microsoft.com/mssql/server:${var.mssql_version}-latest"
  keep_locally = true
}

resource "docker_container" "mssql" {
  image = docker_image.mssql.image_id
  name  = "terraform-mssql-acc-test"

  env = [
    "ACCEPT_EULA=Y",
    "MSSQL_SA_PASSWORD=${random_password.mssql.result}"
  ]

  ports {
    internal = 1433
    external = 11433
  }
}

resource "time_sleep" "mssql_start" {
  create_duration = "5s"

  triggers = {
    container = docker_container.mssql.id
  }
}