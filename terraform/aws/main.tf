variable "name" {
    description = "project name"
    type = string
    default = "mytube"
}

variable "organization" {
  type = string
  default = "51390"
}

variable "docker_host" {
  type = string
  default = "unix:///var/run/docker.sock"
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

provider "docker" {
  host = var.docker_host
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

    enable_nat_gateway = true
    single_nat_gateway = true
    enable_dns_hostnames = true

    public_subnet_tags = {
      "kubernetes.io/cluster/${var.name}" = "shared"
        "kubernetes.io/role/elb"                      = 1
    }

    private_subnet_tags = {
      "kubernetes.io/cluster/${var.name}" = "shared"
        "kubernetes.io/role/internal-elb"             = 1
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

resource "aws_iam_role" "eks_iam_role" {
 name = "devopsthehardway-eks-iam-role"

 path = "/"

 assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
  {
   "Effect": "Allow",
   "Principal": {
    "Service": "eks.amazonaws.com"
   },
   "Action": "sts:AssumeRole"
  }
 ]
}
EOF

}

resource "aws_iam_role_policy_attachment" "AmazonEKSClusterPolicy" {
 policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
 role    = aws_iam_role.eks_iam_role.name
}
resource "aws_iam_role_policy_attachment" "AmazonEC2ContainerRegistryReadOnly-EKS" {
 policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
 role    = aws_iam_role.eks_iam_role.name
}

resource "aws_eks_cluster" "mytube" {
 name = var.name
 role_arn = aws_iam_role.eks_iam_role.arn

 vpc_config {
   subnet_ids = module.vpc.private_subnets
   endpoint_private_access = true
   endpoint_public_access  = true
 }

 depends_on = [
  aws_iam_role.eks_iam_role,
 ]
}

resource "aws_iam_role" "eks_worker_iam_role" {
  name = "eks-node-group-example"

    assume_role_policy = jsonencode({
        Statement = [{
          Action = "sts:AssumeRole"
          Effect = "Allow"
          Principal = {
            Service = "ec2.amazonaws.com"
          }
        }]
        Version = "2012-10-17"
    })
}

 
resource "aws_iam_role_policy_attachment" "AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role    = aws_iam_role.eks_worker_iam_role.name
}

resource "aws_iam_role_policy_attachment" "AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role    = aws_iam_role.eks_worker_iam_role.name
}

resource "aws_iam_role_policy_attachment" "EC2InstanceProfileForImageBuilderECRContainerBuilds" {
  policy_arn = "arn:aws:iam::aws:policy/EC2InstanceProfileForImageBuilderECRContainerBuilds"
  role    = aws_iam_role.eks_worker_iam_role.name
}

resource "aws_iam_role_policy_attachment" "AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role    = aws_iam_role.eks_worker_iam_role.name
}

resource "aws_eks_node_group" "mytube" {
  cluster_name = var.name
  node_group_name = "service"
  subnet_ids = module.vpc.private_subnets
  node_role_arn = aws_iam_role.eks_worker_iam_role.arn
  instance_types = ["t2.micro"]

  scaling_config {
    desired_size = 1
    max_size     = 2
    min_size     = 1
  }

  update_config {
    max_unavailable = 1
  }

  depends_on = [
   aws_iam_role_policy_attachment.AmazonEKSWorkerNodePolicy,
   aws_iam_role_policy_attachment.AmazonEKS_CNI_Policy,
   aws_iam_role_policy_attachment.AmazonEC2ContainerRegistryReadOnly,
   aws_eks_cluster.mytube,
  ]
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
