resource "local_file" "env" {
  filename = "${path.module}/../../local.env"
  content = templatefile("${path.module}/env.tpl", {
    envs = {
      TF_ACC            = 1
      TF_MSSQL_HOST     = "localhost:${docker_container.mssql.ports[0].external}"
      TF_MSSQL_PASSWORD = random_password.mssql.result
      TF_MSSQL_EDITION  = "docker"
    }
  })
}