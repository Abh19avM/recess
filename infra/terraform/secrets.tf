# 1. DATABASE_URL Parameter
resource "aws_ssm_parameter" "db_url" {
  name        = "/${var.app_name}/${var.environment}/DATABASE_URL"
  description = "PostgreSQL connection string for Recess Go backend"
  type        = "SecureString"
  value       = "postgres://${var.db_username}:${var.db_password}@${aws_db_instance.postgres.endpoint}/${aws_db_instance.postgres.db_name}?sslmode=require"

  tags = {
    Name = "${var.app_name}-db-url"
  }
}

# 2. REDIS_URL Parameter
resource "aws_ssm_parameter" "redis_url" {
  name        = "/${var.app_name}/${var.environment}/REDIS_URL"
  description = "Redis connection string for Recess Pub/Sub & state caching"
  type        = "SecureString"
  value       = "redis://${aws_elasticache_cluster.redis.cache_nodes[0].address}:${aws_elasticache_cluster.redis.cache_nodes[0].port}"

  tags = {
    Name = "${var.app_name}-redis-url"
  }
}

# 3. JWT_SECRET Parameter
resource "aws_ssm_parameter" "jwt_secret" {
  name        = "/${var.app_name}/${var.environment}/JWT_SECRET"
  description = "JWT signing key for authentication tokens"
  type        = "SecureString"
  value       = var.jwt_secret

  tags = {
    Name = "${var.app_name}-jwt-secret"
  }
}

# 4. CORS_ALLOWED_ORIGINS Parameter
resource "aws_ssm_parameter" "cors_origins" {
  name        = "/${var.app_name}/${var.environment}/CORS_ALLOWED_ORIGINS"
  description = "Allowed origins for CORS policy"
  type        = "String"
  value       = "https://${aws_cloudfront_distribution.frontend.domain_name}${var.domain_name != "" ? ",https://${var.domain_name}" : ""}"

  tags = {
    Name = "${var.app_name}-cors-origins"
  }
}
