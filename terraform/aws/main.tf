variable "name" {
    description = "name of the application"
    type = string
    default = "mytube"
}

variable "org" {
  type = string
  default = "51390"
}


terraform {
  required_version = ">= 1.2.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
    }
  }
}

provider "aws" {
  region = "sa-east-1"
}

module "vpc" {
    source = "terraform-aws-modules/vpc/aws"
    name = "${var.name}-vpc"
    cidr = "10.0.0.0/24"
    azs = ["sa-east-1a", "sa-east-1b"]
    private_subnets = ["10.0.0.0/26", "10.0.0.64/26"]
    public_subnets = ["10.0.0.128/26", "10.0.0.192/26"]

    tags = {
        Terraform = "true"
        Environment = "dev"
    }
}

#resource "aws_instance" "app_server" {
#  #ami           = "ami-0e2f00f1a5c710177"
#  ami            = "ami-08f7b64eaada185b7"
#  instance_type = "t2.micro"
#
#  tags = {
#    Name = var.app_name
#  }
#}
