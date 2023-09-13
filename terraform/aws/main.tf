variable "app_name" {
    description = "name of the application"
    type = string
    default = "mytube"
}

variable "org" {
  type = string
  default = "51390"
}


terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
  }

  required_version = ">= 1.2.0"

  cloud {
    organization = "51390"
    workspaces {
      name = "mytube"
    }
  }
}

provider "aws" {
  region = "sa-east-1"
}

resource "aws_instance" "app_server" {
  #ami           = "ami-0e2f00f1a5c710177"
  ami            = "ami-08f7b64eaada185b7"
  instance_type = "t2.micro"

  tags = {
    Name = var.app_name
  }
}
