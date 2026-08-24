# CloudWatch Log Group for ECS Backend Task Logs
resource "aws_cloudwatch_log_group" "ecs_backend" {
  name              = "/ecs/${var.app_name}-${var.environment}-backend"
  retention_in_days = 7 # Kept at 7 days to eliminate unnecessary storage cost

  tags = {
    Name = "${var.app_name}-backend-logs"
  }
}

# CloudWatch Alarm for High Memory Utilization (> 85%)
resource "aws_cloudwatch_metric_alarm" "high_memory" {
  alarm_name          = "${var.app_name}-${var.environment}-high-memory"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "MemoryUtilization"
  namespace           = "AWS/ECS"
  period              = 60
  statistic           = "Average"
  threshold           = 85
  alarm_description   = "Alarm when ECS task memory exceeds 85%"

  dimensions = {
    ClusterName = aws_ecs_cluster.main.name
    ServiceName = aws_ecs_service.backend.name
  }

  tags = {
    Name = "${var.app_name}-high-memory-alarm"
  }
}
