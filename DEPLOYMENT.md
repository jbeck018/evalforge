# EvalForge Deployment Guide

## Table of Contents
- [Local Development](#local-development)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Cloud Deployments](#cloud-deployments)
- [Production Checklist](#production-checklist)
- [Monitoring & Maintenance](#monitoring--maintenance)

## Local Development

### Prerequisites
- Docker Desktop or Docker Engine
- Docker Compose v2.0+
- Git
- Python 3.8+ (for testing)
- Node.js 20+ (for frontend development)

### Quick Start

1. **Clone the repository:**
```bash
git clone https://github.com/yourusername/evalforge.git
cd evalforge
```

2. **Set up environment variables:**
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. **Start all services:**
```bash
docker-compose up -d
```

4. **Verify deployment:**
```bash
# Check service health
curl http://localhost:8088/health

# Run production tests
python test_production_readiness.py
```

5. **Access services:**
- Dashboard: http://localhost:3000
- API: http://localhost:8088
- Swagger: http://localhost:8089
- Grafana: http://localhost:3001
- Jaeger: http://localhost:16686

## Docker Deployment

### Building Images

**Build all images:**
```bash
docker-compose build
```

**Build specific service:**
```bash
docker-compose build backend
docker-compose build frontend
```

### Running Services

**Start all services:**
```bash
docker-compose up -d
```

**Start specific services:**
```bash
docker-compose up -d postgres redis
docker-compose up -d backend
docker-compose up -d frontend
```

**View logs:**
```bash
docker-compose logs -f backend
docker-compose logs --tail=100 frontend
```

**Stop services:**
```bash
docker-compose down
# With volume cleanup
docker-compose down -v
```

### Docker Compose Configuration

The `docker-compose.yml` includes:
- **backend**: Go API server (port 8088)
- **frontend**: React dashboard (port 3000)
- **postgres**: PostgreSQL database (port 5432)
- **clickhouse**: Time-series database (ports 8123, 9000)
- **redis**: Cache (port 6379)
- **mock-llm**: Testing LLM (ports 8080-8081)
- **prometheus**: Metrics (port 9090)
- **grafana**: Dashboards (port 3001)
- **jaeger**: Tracing (port 16686)

### Scaling Services

```bash
# Scale backend instances
docker-compose up -d --scale backend=3

# Scale with resource limits
docker-compose --compatibility up -d
```

## Kubernetes Deployment

### Prerequisites
- Kubernetes cluster (1.24+)
- kubectl configured
- Helm 3+ (optional)

### Deploy with kubectl

1. **Create namespace:**
```bash
kubectl create namespace evalforge
```

2. **Create secrets:**
```bash
kubectl create secret generic evalforge-secrets \
  --from-literal=postgres-password=your-password \
  --from-literal=redis-password=your-password \
  --from-literal=jwt-secret=your-jwt-secret \
  -n evalforge
```

3. **Apply configurations:**
```bash
kubectl apply -f k8s/configmap.yaml -n evalforge
kubectl apply -f k8s/postgres.yaml -n evalforge
kubectl apply -f k8s/redis.yaml -n evalforge
kubectl apply -f k8s/backend.yaml -n evalforge
kubectl apply -f k8s/frontend.yaml -n evalforge
kubectl apply -f k8s/ingress.yaml -n evalforge
```

4. **Verify deployment:**
```bash
kubectl get pods -n evalforge
kubectl get services -n evalforge
```

### Deploy with Helm

1. **Add Helm repository:**
```bash
helm repo add evalforge https://charts.evalforge.ai
helm repo update
```

2. **Install chart:**
```bash
helm install evalforge evalforge/evalforge \
  --namespace evalforge \
  --create-namespace \
  --values values.yaml
```

3. **Upgrade deployment:**
```bash
helm upgrade evalforge evalforge/evalforge \
  --namespace evalforge \
  --values values.yaml
```

### Kubernetes Resources

```yaml
# Example backend deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: evalforge-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: evalforge-backend
  template:
    metadata:
      labels:
        app: evalforge-backend
    spec:
      containers:
      - name: backend
        image: evalforge/backend:latest
        ports:
        - containerPort: 8088
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: evalforge-secrets
              key: database-url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

## Cloud Deployments

### AWS Deployment

#### Using ECS

1. **Push images to ECR:**
```bash
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_REGISTRY
docker tag evalforge-backend:latest $ECR_REGISTRY/evalforge-backend:latest
docker push $ECR_REGISTRY/evalforge-backend:latest
```

2. **Create ECS task definition:**
```json
{
  "family": "evalforge-backend",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "backend",
      "image": "your-ecr-registry/evalforge-backend:latest",
      "portMappings": [
        {
          "containerPort": 8088,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "DATABASE_URL",
          "value": "postgres://..."
        }
      ]
    }
  ]
}
```

3. **Create ECS service with ALB**

#### Using EKS

```bash
# Create EKS cluster
eksctl create cluster --name evalforge --region us-east-1

# Deploy application
kubectl apply -f k8s/ -n evalforge
```

#### RDS and ElastiCache

```bash
# Create RDS PostgreSQL
aws rds create-db-instance \
  --db-instance-identifier evalforge-db \
  --db-instance-class db.t3.medium \
  --engine postgres \
  --allocated-storage 100

# Create ElastiCache Redis
aws elasticache create-cache-cluster \
  --cache-cluster-id evalforge-redis \
  --cache-node-type cache.t3.micro \
  --engine redis
```

### Google Cloud Deployment

#### Using Cloud Run

```bash
# Build and push to Container Registry
gcloud builds submit --tag gcr.io/PROJECT-ID/evalforge-backend

# Deploy to Cloud Run
gcloud run deploy evalforge-backend \
  --image gcr.io/PROJECT-ID/evalforge-backend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars DATABASE_URL=postgres://...
```

#### Using GKE

```bash
# Create GKE cluster
gcloud container clusters create evalforge \
  --zone us-central1-a \
  --num-nodes 3

# Deploy application
kubectl apply -f k8s/ -n evalforge
```

### Azure Deployment

#### Using Container Instances

```bash
# Create container group
az container create \
  --resource-group evalforge-rg \
  --name evalforge-backend \
  --image evalforge/backend:latest \
  --dns-name-label evalforge \
  --ports 8088 \
  --environment-variables DATABASE_URL=postgres://...
```

#### Using AKS

```bash
# Create AKS cluster
az aks create \
  --resource-group evalforge-rg \
  --name evalforge-cluster \
  --node-count 3

# Deploy application
kubectl apply -f k8s/ -n evalforge
```

## Production Checklist

### Security

- [ ] Enable HTTPS/TLS certificates
- [ ] Configure firewall rules
- [ ] Set up API rate limiting
- [ ] Enable audit logging
- [ ] Rotate secrets regularly
- [ ] Configure CORS properly
- [ ] Set secure headers (HSTS, CSP)
- [ ] Enable database encryption
- [ ] Set up VPN/private networking
- [ ] Configure WAF rules

### Database

- [ ] Set up database backups
- [ ] Configure replication
- [ ] Optimize connection pooling
- [ ] Create read replicas
- [ ] Set up monitoring alerts
- [ ] Configure auto-scaling
- [ ] Optimize indexes
- [ ] Set up maintenance windows

### Application

- [ ] Configure health checks
- [ ] Set up auto-scaling
- [ ] Configure load balancing
- [ ] Enable distributed tracing
- [ ] Set up error tracking
- [ ] Configure log aggregation
- [ ] Set up CDN for static assets
- [ ] Enable caching strategies
- [ ] Configure rate limiting
- [ ] Set up circuit breakers

### Monitoring

- [ ] Configure Prometheus metrics
- [ ] Set up Grafana dashboards
- [ ] Configure alerting rules
- [ ] Set up PagerDuty integration
- [ ] Configure log aggregation
- [ ] Set up uptime monitoring
- [ ] Configure APM tools
- [ ] Set up cost monitoring

## Monitoring & Maintenance

### Health Checks

```bash
# API health
curl http://your-domain.com/health

# Database health
docker exec evalforge_postgres pg_isready

# Redis health
docker exec evalforge_redis redis-cli ping
```

### Backup & Restore

#### PostgreSQL Backup

```bash
# Backup
docker exec evalforge_postgres pg_dump -U evalforge evalforge > backup.sql

# Restore
docker exec -i evalforge_postgres psql -U evalforge evalforge < backup.sql
```

#### ClickHouse Backup

```bash
# Backup
docker exec evalforge_clickhouse clickhouse-backup create

# Restore
docker exec evalforge_clickhouse clickhouse-backup restore backup_name
```

### Log Management

```bash
# View logs
docker-compose logs -f --tail=100

# Export logs
docker-compose logs > evalforge.log

# Log rotation
cat > /etc/logrotate.d/evalforge << EOF
/var/log/evalforge/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 evalforge evalforge
    sharedscripts
}
EOF
```

### Performance Tuning

#### PostgreSQL

```sql
-- Optimize connections
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';

-- Create indexes
CREATE INDEX idx_events_project_time ON trace_events(project_id, created_at);
CREATE INDEX idx_evaluations_project ON evaluations(project_id);
```

#### Backend

```yaml
# Environment variables for tuning
GOMAXPROCS: 4
GIN_MODE: release
DATABASE_MAX_CONNECTIONS: 100
DATABASE_MAX_IDLE_CONNECTIONS: 10
REDIS_POOL_SIZE: 50
```

### Updates & Upgrades

```bash
# Pull latest images
docker-compose pull

# Update with zero downtime
docker-compose up -d --no-deps --build backend
docker-compose up -d --no-deps --build frontend

# Database migrations
docker exec evalforge_backend /app/migrate up
```

### Troubleshooting

#### Common Issues

1. **Database connection errors:**
```bash
# Check database status
docker-compose ps postgres
docker logs evalforge_postgres

# Test connection
docker exec evalforge_postgres psql -U evalforge -c "SELECT 1"
```

2. **High memory usage:**
```bash
# Check memory usage
docker stats

# Restart with memory limits
docker-compose down
docker-compose up -d --compatibility
```

3. **Slow API responses:**
```bash
# Check slow queries
docker exec evalforge_postgres psql -U evalforge -c "SELECT * FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10"

# Check backend metrics
curl http://localhost:8088/metrics | grep http_request_duration
```

## Support

For deployment assistance:
- Documentation: https://docs.evalforge.ai/deployment
- Community: https://discord.gg/evalforge
- Issues: https://github.com/yourusername/evalforge/issues