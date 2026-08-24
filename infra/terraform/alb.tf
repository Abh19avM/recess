# Application Load Balancer (ALB)
resource "aws_lb" "main" {
  name               = "${var.app_name}-${var.environment}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = [aws_subnet.public_1.id, aws_subnet.public_2.id]

  # 300s idle timeout to maintain active long-lived WebSocket duels without timeouts
  idle_timeout = 300

  tags = {
    Name = "${var.app_name}-${var.environment}-alb"
  }
}

# Target Group pointing to Go backend ECS Fargate tasks on port 8080
resource "aws_lb_target_group" "backend" {
  name        = "${var.app_name}-${var.environment}-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    enabled             = true
    path                = "/health"
    protocol            = "HTTP"
    port                = "8080"
    interval            = 15
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
    matcher             = "200"
  }

  tags = {
    Name = "${var.app_name}-${var.environment}-tg"
  }
}

# HTTP Listener (Port 80) -> Redirect to HTTPS if domain is set, or forward to TG
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.backend.arn
  }
}
