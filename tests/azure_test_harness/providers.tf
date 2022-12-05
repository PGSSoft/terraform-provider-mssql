terraform {
  required_providers {
    publicip = {
      source = "nxt-engineering/publicip"
    }

    environment = {
      source = "EppO/environment"
    }
  }
}

provider "azurerm" {
  features {}
}

