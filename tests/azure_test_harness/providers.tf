terraform {
  required_providers {
    publicip = {
      source = "nxt-engineering/publicip"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "publicip" {}