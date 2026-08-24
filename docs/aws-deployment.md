# AWS Production Deployment Guide

This guide covers provisioning, configuring, deploying, monitoring, and rolling back **Recess** on Amazon Web Services (AWS) using Terraform, ECS Fargate, RDS PostgreSQL, ElastiCache Redis, S3, and CloudFront.

---

## 1. Target AWS Architecture

```
                       [ Internet Users ]
                         │            │
         Static Assets / │            │ API & WebSockets
           SPA Routes    │            │ (/api/*, /ws*)
                         ▼            ▼
             ┌─────────────────┐  ┌─────────────────┐
             │ CloudFront CDN  │  │ Application LB  │
             │ (HTTPS / OAC)   │  │ (HTTPS / 300s)  │
             └────────┬────────┘  └────────┬────────┘
                      │                    │
                      ▼ (OAC)              ▼ (Port 8080)
             ┌─────────────────┐  ┌───────────────────────┐
             │ S3 Frontend     │  │ ECS Fargate Tasks     │
             │ (Vite React)    │  │ (Go Recess Backend)   │
             └─────────────────┘  └───┬───────────────┬───┘
                                      │               │
                             Postgres │               │ Redis Pub/Sub
                            (Port 5432)               │ (Port 6379)
                                      ▼               ▼
                            ┌────────────────┐ ┌────────────────┐
                            │ AWS RDS        │ │ ElastiCache    │
                            │ PostgreSQL     │ │ Redis          │
                            │ (db.t4g.micro) │ │ (cache.t4g)    │
                            └────────────────┘ └────────────────┘
```

---

## 2. Cost Breakdown & AWS Credit Optimization

To protect limited AWS credits and stay within Free Tier limits, resources have been sized for maximum cost efficiency:

| Resource | Sizing & Configuration | Est. Monthly Cost | Optimization Strategy |
| :--- | :--- | :--- | :--- |
| **CloudFront CDN** | PriceClass_100 (NA/Europe), OAC | **~$0.00** (Free Tier) | 1 TB data transfer free per month. |
| **S3 Storage** | Static SPA build (< 50 MB) | **~$0.10** | Standard storage with automated build sync (`--delete`). |
| **Application Load Balancer** | 1 Regional ALB with 300s timeout | **~$16.00** | Shared listener for REST API and WebSocket paths. |
| **ECS Fargate** | 0.25 vCPU, 512 MB RAM (1 min, 3 max) | **~$7.50 - $15.00** | Autoscales based on CPU/memory utilization (>70%). |
| **Amazon RDS PostgreSQL** | `db.t4g.micro`, 20 GB gp3, Single-AZ | **~$12.50** (Free Tier eligible) | Single-AZ burstable Graviton instance with 7-day backups. |
| **ElastiCache Redis** | `cache.t4g.micro`, 1 node | **~$11.00** | Single-node cache for low-latency Pub/Sub and session state. |
| **Amazon ECR** | 1 Private Repository | **~$0.10** | Lifecycle policy retaining only the 5 most recent image tags. |
| **AWS SSM Parameter Store** | Standard SecureString parameters | **$0.00** (Free) | Zero cost configuration and credential injection. |
| **Amazon CloudWatch** | ECS Log Group (7-day retention) | **~$1.00** | Retains logs for 7 days only to prevent runaway storage bills. |
| **Total Estimated Cost** | | **~$48 - $56 / month** | *Over 85% cost reduction vs multi-AZ enterprise setups ($350+/mo).* |

---

## 3. Pre-Requisites & IAM Setup

1. **AWS CLI** installed and configured:
   ```bash
   aws configure
   ```
2. **Terraform** (`>= 1.5.0`):
   ```bash
   terraform version
   ```
3. **AWS IAM User / Role** with permissions for:
   - `AmazonVPCFullAccess`, `AmazonECS_FullAccess`, `AmazonRDSFullAccess`
   - `AmazonElastiCacheFullAccess`, `AmazonEC2ContainerRegistryFullAccess`
   - `CloudFrontFullAccess`, `AmazonS3FullAccess`, `AmazonSSMFullAccess`

---

## 4. Step-by-Step Provisioning with Terraform

### Step 1: Initialize Configuration
```bash
cd infra/terraform
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars`:
```hcl
aws_region  = "us-east-1"
environment = "production"
app_name    = "recess"

db_username = "recess_admin"
db_password = "MySecurePassword123!" # Change to strong password
jwt_secret  = "super-secure-jwt-secret-key-32-chars-long"
```

### Step 2: Apply Terraform Plan
```bash
terraform init
terraform plan -out=tfplan
terraform apply tfplan
```

### Step 3: Note Outputs
After provisioning completes, note the outputs:
- `ecr_repository_url`
- `alb_dns_name`
- `cloudfront_domain_name`
- `s3_bucket_name`

---

## 5. Build and Deploy Containers

### 1. Build and Push Go Backend to ECR
```bash
# Authenticate Docker to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <ECR_URL>

# Build image
docker build -t <ECR_URL>:latest -f backend/Dockerfile backend/

# Push image
docker push <ECR_URL>:latest

# Trigger ECS Service Update
aws ecs update-service --cluster recess-production-cluster --service recess-backend-service --force-new-deployment
```

### 2. Build and Deploy React Frontend to S3 & CloudFront
```bash
cd frontend

# Set API & WebSocket URLs in production build
VITE_API_URL=http://<ALB_DNS_NAME> \
VITE_WS_URL=ws://<ALB_DNS_NAME>/ws \
npm run build

# Sync dist directory to S3
aws s3 sync dist/ s3://<S3_BUCKET_NAME>/ --delete

# Invalidate CloudFront cache
aws cloudfront create-invalidation --distribution-id <DISTRIBUTION_ID> --paths "/*"
```

---

## 6. Database Migration on RDS

Run initial migrations against the RDS instance (from an EC2 bastion host, ECS Task run, or local CLI through an authorized IP):
```bash
DATABASE_URL="postgres://recess_admin:MySecurePassword123!@<RDS_ENDPOINT>/recess?sslmode=require" \
go run ./cmd/migrate up
```

---

## 7. Verification & Smoke Testing Checklist

- [ ] **Health Check**: `curl -s http://<ALB_DNS_NAME>/health` returns `{"status":"ok"}`.
- [ ] **REST API**: `curl -s http://<ALB_DNS_NAME>/api/v1/rooms` returns active rooms payload.
- [ ] **SPA Navigation**: Visiting `https://<CLOUDFRONT_DOMAIN>/dashboard` loads without 404.
- [ ] **Guest Login**: Clicking "Quick Guest Play" generates valid JWT access & refresh tokens.
- [ ] **WebSocket Connection**: Real-time handshake over `ws://<ALB_DNS_NAME>/ws?token=...` establishes and responds to `room.ping`.
- [ ] **Multiplayer Matchmaking**: 2 browsers can pair in Hand Cricket, Dots & Boxes, XO, Connect 4, Paper Football, and NPAT.
- [ ] **Spectator Mode**: Third spectator client receives live game state updates without submitting moves.

---

## 8. Rollback Runbook (Zero Downtime)

### Rollback Backend (ECS Fargate)
If a bad release is deployed to ECS, revert to the previous task definition revision immediately:

```bash
# 1. List recent task definition revisions
aws ecs list-task-definitions --family-prefix recess-backend --sort DESC

# 2. Update service to point to previous revision (e.g. revision 2)
aws ecs update-service \
  --cluster recess-production-cluster \
  --service recess-backend-service \
  --task-definition recess-backend:2

# 3. Monitor ECS deployment event stream
aws ecs describe-services --cluster recess-production-cluster --services recess-backend-service --query "services[0].events[:5]"
```

### Rollback Frontend (S3 / CloudFront)
```bash
# Re-upload previous build artifact
aws s3 sync ./backup-dist/ s3://<S3_BUCKET_NAME>/ --delete

# Invalidate CloudFront cache
aws cloudfront create-invalidation --distribution-id <DISTRIBUTION_ID> --paths "/*"
```

---

## 9. Budget Alerts & Cost Protection

To ensure AWS spending never exceeds your credit threshold:
1. Navigate to **AWS Billing -> Budgets**.
2. Create a **Monthly Cost Budget** with threshold **$50.00**.
3. Set email alerts at 80% ($40) and 100% ($50) utilization.
