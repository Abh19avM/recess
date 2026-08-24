variable "aws_region" {
  description = "AWS deployment region (e.g. us-east-1, ap-south-1, eu-west-1)"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Application deployment environment"
  type        = string
  default     = "production"
}

variable "app_name" {
  description = "Name of the application"
  type        = string
  default     = "recess"
}

variable "domain_name" {
  description = "Root domain name for the application (leave empty if using direct ALB/CloudFront domains)"
  type        = string
  default     = ""
}

variable "db_username" {
  description = "Master username for PostgreSQL database"
  type        = string
  default     = "recess_admin"
}

variable "db_password" {
  description = "Master password for PostgreSQL database (must be at least 12 chars)"
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "Cryptographic secret key for signing JWT tokens (HMAC-SHA256)"
  type        = string
  sensitive   = true
}

variable "container_image" {
  description = "Docker image URI in ECR for the Go backend"
  type        = string
  default     = ""
}

variable "fargate_cpu" {
  description = "Fargate CPU units (256 = 0.25 vCPU, 512 = 0.5 vCPU)"
  type        = number
  default     = 256
}

variable "fargate_memory" {
  description = "Fargate memory in MB (512 = 0.5 GB, 1024 = 1 GB)"
  type        = number
  default     = 512
}

variable "min_capacity" {
  description = "Minimum ECS task count for autoscaling"
  type        = number
  default     = 1
}

variable "max_capacity" {
  description = "Maximum ECS task count for autoscaling"
  type        = number
  default     = 3
}
