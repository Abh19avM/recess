# DB Subnet Group (Private Subnets)
resource "aws_db_subnet_group" "main" {
  name        = "${var.app_name}-${var.environment}-db-subnet-group"
  description = "Subnet group for Recess RDS PostgreSQL database"
  subnet_ids  = [aws_subnet.private_1.id, aws_subnet.private_2.id]

  tags = {
    Name = "${var.app_name}-${var.environment}-db-subnet-group"
  }
}

# PostgreSQL Database (Single-AZ db.t4g.micro for cost optimization)
resource "aws_db_instance" "postgres" {
  identifier             = "${var.app_name}-${var.environment}-db"
  allocated_storage      = 20
  max_allocated_storage  = 50
  storage_type           = "gp3"
  engine                 = "postgres"
  engine_version         = "16.2"
  instance_class         = "db.t4g.micro"
  db_name                = "recess"
  username               = var.db_username
  password               = var.db_password
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false
  skip_final_snapshot    = true
  deletion_protection    = false

  backup_retention_period   = 7
  auto_minor_version_upgrade = true

  tags = {
    Name = "${var.app_name}-${var.environment}-postgres"
  }
}
