locals {
  name = "tfmsqltest${random_string.suffix.result}"
}

resource "random_string" "suffix" {
  length  = 5
  special = false
  upper   = false
}

data "environment_variables" "arm" {
  filter = "^ARM_"
}