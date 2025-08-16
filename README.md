---

````markdown
# 📬 AdmailPro Backend

AdmailPro Backend is a high-performance email campaign platform built with **Go (fasthttp)**. It supports features like user authentication, CSV uploads, campaign creation, subdomain management, and more. It is containerized using Docker and ready to deploy using GitLab CI/CD and AWS Lambda/API Gateway.

---

##  Quick Start with Docker

###  Prerequisites

- Docker & Docker Compose installed
- Go 1.22+ installed (for local builds, optional)

###  Build and Run All Services

From the root project folder (`email-sender/`), run:

```bash
docker compose up --build
````

This spins up:

| Service | Description          | Port    |
| ------- | -------------------- | ------- |
| Backend | Go API with fasthttp | `8080`  |
| MongoDB | Document database    | `27017` |
| Redis   | Background job queue | `6379`  |

---

##  Health Check

Verify backend is running:

```bash
curl http://localhost:8080/healthz
```

Expected response:

```
OK
```

---

## ⚙️ Environment Variables

These are passed in `docker-compose.yml`:

```env
MONGO_URI=mongodb://mongo:27017
REDIS_ADDR=redis:6379
JWT_SECRET=super-secret-jwt-key
AWS_ACCESS_KEY=your-access-key
AWS_SECRET_KEY=your-secret-key
AWS_REGION=us-east-1
TRACKING_DOMAIN=track.localhost
```

---

##  Sample API Testing

### Signup

```bash
curl -X POST http://localhost:8080/api/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com", "password":"secure123"}'
```

### 🔑 Login

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com", "password":"secure123"}'
```

---

## Logs & Debugging

### View Logs

```bash
docker compose logs -f backend
```

### View All Logs

```bash
docker compose logs -f
```

---

## Interact with Containers

### List Running Containers

```bash
docker ps
```

### Access Mongo Shell

```bash
docker compose exec mongo /bin/bash
mongosh
```

*If `mongosh` is unavailable, try `mongo`.*

---

## Manual Docker Build (Optional)

```bash
docker build -t admail-backend -f Dockerfile .
docker run -p 8080:8080 admail-backend
```

---

## Authors & Acknowledgments

Built by the AdmailPro backend team. Inspired by fast, minimal Go practices.

---

##  License

This project is licensed under the MIT License. See `LICENSE` for full text.


---
