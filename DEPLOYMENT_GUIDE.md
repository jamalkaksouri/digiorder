# Production Deployment Guide

## Overview

This guide covers deploying DigiOrder v2.0 to production using Docker, with CI/CD via GitHub Actions.

---

## Prerequisites

- Docker & Docker Compose installed
- GitHub account for CI/CD
- Domain name (optional)
- SSL certificate (recommended)
- Minimum server specs:
  - 2 CPU cores
  - 4GB RAM
  - 20GB storage

---

## Quick Start (Docker)

### 1. Clone Repository

```bash
git clone https://github.com/jamalkaksouri/DigiOrder.git
cd DigiOrder
```

### 2. Configure Environment

```bash
cp .env.production .env
```

Edit `.env` and set:

```bash
# Database
DB_PASSWORD=YOUR_STRONG_PASSWORD_HERE
JWT_SECRET=YOUR_RANDOM_SECRET_64_CHARS
```

### 3. Generate Strong Secrets

```bash
# Generate JWT secret
openssl rand -base64 64

# Generate DB password
pwgen -s 32 1
```

### 4. Deploy

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### 5. Verify

```bash
# Check containers
docker-compose -f docker-compose.prod.yml ps

# Check health
curl http://localhost:5582/health
```

---

## Step-by-Step Production Deployment

### 1. Server Setup

#### Update System

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y docker.io docker-compose git make
```

#### Configure Firewall

```bash
# Allow SSH, HTTP, HTTPS
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 2. Install Docker

```bash
# Add Docker's official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Set up stable repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt update
sudo apt install docker-ce docker-ce-cli containerd.io docker-compose-plugin -y

# Add user to docker group
sudo usermod -aG docker $USER
```

### 3. Application Deployment

```bash
# Create application directory
sudo mkdir -p /opt/digiorder
sudo chown $USER:$USER /opt/digiorder
cd /opt/digiorder

# Clone repository
git clone https://github.com/jamalkaksouri/DigiOrder.git .

# Configure environment
cp .env.production .env

# Edit configuration
nano .env
```

#### Required Configuration

```bash
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=digiorder_prod
DB_PASSWORD=<GENERATE_STRONG_PASSWORD>
DB_NAME=digiorder_production
DB_SSLMODE=require

# Server Configuration
SERVER_PORT=5582
SERVER_HOST=0.0.0.0

# JWT Configuration
JWT_SECRET=<GENERATE_RANDOM_64_CHAR_SECRET>
JWT_EXPIRY=24h

# CORS (adjust for your domain)
CORS_ALLOWED_ORIGINS=https://yourdomain.com
```

### 4. Build and Deploy

```bash
# Build image
docker-compose -f docker-compose.prod.yml build

# Start services
docker-compose -f docker-compose.prod.yml up -d

# Check logs
docker-compose -f docker-compose.prod.yml logs -f
```

### 5. Create Initial Admin User

```bash
# Connect to database
docker-compose -f docker-compose.prod.yml exec postgres psql -U digiorder_prod -d digiorder_production

# Create admin user (password hash for "admin123456")
INSERT INTO users (username, full_name, password_hash, role_id)
VALUES (
  'admin',
  'System Administrator',
  '$2a$10$Zu7yVNJ0e9Fn9vwUy9vRbO5CqPQZMB8l5k8hEWnGvhkrFUKqj9iEW',
  1
);
```

Or use this script:

```bash
docker-compose -f docker-compose.prod.yml exec -T postgres psql -U digiorder_prod -d digiorder_production <<EOF
INSERT INTO users (username, full_name, password_hash, role_id)
VALUES ('admin', 'System Administrator', '\$2a\$10\$Zu7yVNJ0e9Fn9vwUy9vRbO5CqPQZMB8l5k8hEWnGvhkrFUKqj9iEW', 1)
ON CONFLICT (username) DO NOTHING;
EOF
```

---

## Nginx Reverse Proxy Setup

### 1. Install Nginx

```bash
sudo apt install nginx -y
```

### 2. Configure Nginx

Create `/etc/nginx/sites-available/digiorder`:

```nginx
upstream digiorder_backend {
    server localhost:5582;
}

server {
    listen 80;
    server_name api.yourdomain.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    # SSL Configuration
    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Logging
    access_log /var/log/nginx/digiorder_access.log;
    error_log /var/log/nginx/digiorder_error.log;

    # Proxy settings
    location / {
        proxy_pass http://digiorder_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Health check endpoint (no auth required)
    location /health {
        proxy_pass http://digiorder_backend;
        access_log off;
    }

    # Static files (if any)
    location /static/ {
        alias /opt/digiorder/static/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
```

### 3. Enable Site

```bash
sudo ln -s /etc/nginx/sites-available/digiorder /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 4. SSL Certificate (Let's Encrypt)

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx -y

# Obtain certificate
sudo certbot --nginx -d api.yourdomain.com

# Auto-renewal is configured automatically
sudo certbot renew --dry-run
```

---

## CI/CD with GitHub Actions

### 1. Setup GitHub Secrets

Go to GitHub repo → Settings → Secrets and variables → Actions

Add these secrets:

```
DOCKER_USERNAME=your_dockerhub_username
DOCKER_PASSWORD=your_dockerhub_password
DEPLOY_HOST=your.server.ip
DEPLOY_USER=your_ssh_user
DEPLOY_KEY=<your_ssh_private_key>
```

### 2. Generate SSH Key

```bash
# On your local machine
ssh-keygen -t ed25519 -C "github-actions@digiorder"

# Copy public key to server
ssh-copy-id -i ~/.ssh/id_ed25519.pub user@your.server.ip

# Copy private key content for GitHub secret
cat ~/.ssh/id_ed25519
```

### 3. Workflow is Auto-Triggered

The `.github/workflows/ci.yml` file will:

1. Run tests on push
2. Build Docker image
3. Push to Docker Hub
4. Deploy to production server

---

## Monitoring & Maintenance

### 1. View Logs

```bash
# All services
docker-compose -f docker-compose.prod.yml logs -f

# Specific service
docker-compose -f docker-compose.prod.yml logs -f api

# Last 100 lines
docker-compose -f docker-compose.prod.yml logs --tail=100
```

### 2. Monitor Resources

```bash
# Container stats
docker stats

# System resources
htop
```

### 3. Database Backup

```bash
# Create backup directory
mkdir -p /opt/digiorder/backups

# Backup script
cat > /opt/digiorder/backup.sh <<'EOF'
#!/bin/bash
BACKUP_DIR="/opt/digiorder/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/digiorder_backup_$DATE.sql"

docker-compose -f /opt/digiorder/docker-compose.prod.yml exec -T postgres \
  pg_dump -U digiorder_prod digiorder_production > "$BACKUP_FILE"

gzip "$BACKUP_FILE"

# Keep only last 30 days
find $BACKUP_DIR -name "*.sql.gz" -mtime +30 -delete

echo "Backup completed: ${BACKUP_FILE}.gz"
EOF

chmod +x /opt/digiorder/backup.sh
```

### 4. Automated Backups (Cron)

```bash
# Add to crontab
crontab -e

# Daily backup at 2 AM
0 2 * * * /opt/digiorder/backup.sh >> /var/log/digiorder_backup.log 2>&1
```

### 5. Restore from Backup

```bash
# Unzip backup
gunzip digiorder_backup_20250108_020000.sql.gz

# Restore
docker-compose -f docker-compose.prod.yml exec -T postgres \
  psql -U digiorder_prod -d digiorder_production < digiorder_backup_20250108_020000.sql
```

---

## Scaling & Performance

### 1. Horizontal Scaling (Multiple Instances)

Update `docker-compose.prod.yml`:

```yaml
services:
  api:
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
```

### 2. Load Balancer (Nginx)

```nginx
upstream digiorder_backend {
    least_conn;
    server localhost:5582;
    server localhost:5583;
    server localhost:5584;
}
```

### 3. Database Connection Pooling

In `.env`:

```bash
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
```

---

## Security Hardening

### 1. Firewall Rules

```bash
# Only allow necessary ports
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow 22/tcp  # SSH
sudo ufw allow 80/tcp  # HTTP
sudo ufw allow 443/tcp # HTTPS
sudo ufw enable
```

### 2. Fail2Ban (SSH Protection)

```bash
sudo apt install fail2ban -y

# Configure
sudo cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

### 3. Regular Updates

```bash
# Create update script
cat > /opt/digiorder/update.sh <<'EOF'
#!/bin/bash
cd /opt/digiorder
git pull
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d --build
docker system prune -f
EOF

chmod +x /opt/digiorder/update.sh
```

---

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker-compose -f docker-compose.prod.yml logs api

# Check container status
docker-compose -f docker-compose.prod.yml ps

# Restart services
docker-compose -f docker-compose.prod.yml restart
```

### Database Connection Issues

```bash
# Test database connection
docker-compose -f docker-compose.prod.yml exec postgres psql -U digiorder_prod -d digiorder_production -c "SELECT 1;"

# Check database logs
docker-compose -f docker-compose.prod.yml logs postgres
```

### High Memory Usage

```bash
# Check container stats
docker stats

# Restart container
docker-compose -f docker-compose.prod.yml restart api

# Clear Docker cache
docker system prune -a
```

---

## Rollback Procedure

```bash
# Stop current version
docker-compose -f docker-compose.prod.yml down

# Checkout previous version
git log --oneline  # Find commit hash
git checkout <previous_commit_hash>

# Rebuild and deploy
docker-compose -f docker-compose.prod.yml up -d --build
```

---

## Health Checks

### 1. Application Health

```bash
curl https://api.yourdomain.com/health
```

### 2. Automated Monitoring

```bash
# Create monitor script
cat > /opt/digiorder/monitor.sh <<'EOF'
#!/bin/bash
URL="https://api.yourdomain.com/health"
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $URL)

if [ $RESPONSE -ne 200 ]; then
    echo "Health check failed! Status: $RESPONSE"
    # Send alert (email, Slack, etc.)
    # Restart service if needed
    docker-compose -f /opt/digiorder/docker-compose.prod.yml restart api
fi
EOF

chmod +x /opt/digiorder/monitor.sh

# Add to crontab (every 5 minutes)
*/5 * * * * /opt/digiorder/monitor.sh
```

---

## Summary Checklist

- [ ] Server configured with Docker
- [ ] Application deployed via docker-compose
- [ ] Environment variables configured securely
- [ ] Initial admin user created
- [ ] Nginx reverse proxy configured
- [ ] SSL certificate installed
- [ ] Firewall rules applied
- [ ] Automated backups configured
- [ ] Monitoring setup
- [ ] CI/CD pipeline connected
- [ ] Health checks passing
- [ ] Documentation reviewed

---

## Support & Resources

- **Documentation**: https://github.com/jamalkaksouri/DigiOrder/wiki
- **Issues**: https://github.com/jamalkaksouri/DigiOrder/issues
- **Docker Hub**: https://hub.docker.com/r/yourusername/digiorder

---

Your DigiOrder application is now production-ready! 🚀
