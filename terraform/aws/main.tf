variable "name" {
    description = "project name"
    type = string
    default = "mytube"
}

variable "organization" {
  type = string
  default = "51390"
}

variable "region" {
  type = string
  default = "sa-east-1"
}

variable "azs" {
    description = "availability zones"
    type = list(string)
    default = ["sa-east-1a", "sa-east-1b"]
}

locals {
    docker_login_cmd = "aws ecr get-login-password --region ${var.region}"
}

terraform {
  required_version = ">= 1.2.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
    }

    docker = {
      source  = "kreuzwerker/docker"
        version = "3.0.2"
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
    azs = var.azs
    private_subnets = ["10.0.0.0/26", "10.0.0.64/26"]
    public_subnets = ["10.0.0.128/26", "10.0.0.192/26"]

    tags = {
        Terraform = "true"
        Environment = "dev"
    }
}

resource "aws_ecr_repository" "app" {
    name = "${var.name}-app"
    force_delete = true

    image_scanning_configuration {
        scan_on_push = true
    }
}

data "docker_image" "app" {
  name = "${var.name}-app"
}

resource "null_resource" "app_image" {
  triggers = {
    id = data.docker_image.app.id
  }
}

resource "docker_tag" "app" {
  source_image = "${var.name}-app"
  target_image = "${aws_ecr_repository.app.repository_url}"

  lifecycle {
    replace_triggered_by = [
      null_resource.app_image.id
    ]
  }
}

resource "aws_ecr_repository" "service" {
    name = "${var.name}-service"
    force_delete = true

    image_scanning_configuration {
        scan_on_push = true
    }
}

data "docker_image" "service" {
  name = "${var.name}-service"
}

resource "null_resource" "service_image" {
  triggers = {
    id = data.docker_image.service.id
  }
}

resource "docker_tag" "service" {
  source_image = "${var.name}-service"
  target_image = "${aws_ecr_repository.service.repository_url}"

  lifecycle {
    replace_triggered_by = [
      null_resource.service_image.id
    ]
  }
}

resource "aws_ecr_repository" "db" {
    name = "${var.name}-db"
    force_delete = true

    image_scanning_configuration {
        scan_on_push = true
    }
}

data "docker_image" "db" {
  name = "${var.name}-db"
}

resource "null_resource" "db_image" {
  triggers = {
    id = data.docker_image.db.id
  }
}


resource "docker_tag" "db" {
  source_image = "${var.name}-db"
  target_image = "${aws_ecr_repository.db.repository_url}"

  lifecycle {
    replace_triggered_by = [
      null_resource.db_image.id
    ]
  }
}

resource "null_resource" "push_images" {
  triggers = {
    app_image_id = docker_tag.app.source_image_id
    service_image_id = docker_tag.service.source_image_id
    db_image_id = docker_tag.db.source_image_id
  }

  provisioner "local-exec" {
    command = <<EOT
      ${local.docker_login_cmd} | docker login --username AWS --password-stdin ${aws_ecr_repository.app.repository_url}
      docker push ${aws_ecr_repository.app.repository_url}

      ${local.docker_login_cmd} | docker login --username AWS --password-stdin ${aws_ecr_repository.service.repository_url}
      docker push ${aws_ecr_repository.service.repository_url}

      ${local.docker_login_cmd} | docker login --username AWS --password-stdin ${aws_ecr_repository.db.repository_url}
      docker push ${aws_ecr_repository.db.repository_url}
    EOT
  }
}

output "app_repository_url" {
  value = aws_ecr_repository.app.repository_url
}

output "service_repository_url" {
  value = aws_ecr_repository.service.repository_url
}

output "db_repository_url" {
  value = aws_ecr_repository.db.repository_url
}

output "app_image_id" {
  value = data.docker_image.app.id
}

output "service_image_id" {
  value = data.docker_image.service.id
}

output "db_image_id" {
  value = data.docker_image.db.id
}
