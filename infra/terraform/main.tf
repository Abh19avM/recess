terraform {
  required_version = ">= 1.5.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.40"
    }
  }

  # Optional remote state backend (e.g. S3 + DynamoDB for state locking)
  # backend "s3" {
  #   bucket         = "recess-terraform-state"
  #   key            = "prod/terraform.tfstate"
  #   region         = "us-east-1"
  #   dynamodb_table = "recess-terraform-locks"
  #   encrypt        = true
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "Recess"
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}

# Secondary provider for CloudFront ACM certificates (must be in us-east-1)
provider "aws" {
  alias  = "us_east_1"
  region = "us-east-1"

  default_tags {
    tags = {
      Project     = "Recess"
      Environment = var.environment
      ManagedBy   = "Terraform"
    }
  }
}
