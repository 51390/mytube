variable "region" {
    description = "azure region to deploy"
    type = string
    default = "brazilsouth"
}

# Configure the Azure provider
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0.2"
    }
  }

  required_version = ">= 1.1.0"
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "rg" {
  name     = "MyTubeResourceGroup"
  location = var.region
}

# Create a virtual network
resource "azurerm_virtual_network" "vnet" {
  name                = "MyTubeNetwork"
  address_space       = ["10.0.0.0/16"]
  location            = var.region
  resource_group_name = azurerm_resource_group.rg.name
}

