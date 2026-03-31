# School Management System — Local Setup Guide

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (running)
- [curl](https://curl.se/) (for testing APIs)

---

## 🚀 Start All Services

```bash
docker compose up --build -d
```

This starts **4 containers**:

| Container        | Purpose                  | Port  |
| ---------------- | ------------------------ | ----- |
| `auth-db`        | PostgreSQL (Auth)        | 5433  |
| `user-db`        | PostgreSQL (User)        | 5434  |
| `auth-service`   | Auth & User Mgmt API     | 8081  |
| `user-service`   | Roles & Permissions API  | 8082  |

### Verify everything is running

```bash
docker compose ps
curl http://localhost:8081/health
curl http://localhost:8082/health
```

---

## 🛑 Stop / Reset

```bash
# Stop containers (data preserved)
docker compose down

# Stop AND wipe all data (fresh start)
docker compose down -v
```

---

## 📋 API Reference

### Auth Service (`localhost:8081`)

| Method   | Endpoint                | Auth         | Description                        |
| -------- | ----------------------- | ------------ | ---------------------------------- |
| `POST`   | `/auth/register-school` | Public       | Register school + super admin      |
| `POST`   | `/auth/login`           | Public       | Login → JWT token                  |
| `GET`    | `/auth/me`              | Bearer token | Get current user profile           |
| `POST`   | `/users`                | Admin only   | Create user (teacher/staff/parent) |
| `GET`    | `/users`                | Admin only   | List users (filter + paginate)     |
| `GET`    | `/users/:id`            | Admin only   | Get single user                    |
| `PATCH`  | `/users/:id`            | Admin only   | Update user                        |
| `DELETE` | `/users/:id`            | Admin only   | Delete user                        |

### User Service (`localhost:8082`)

| Method | Endpoint                           | Auth         | Description                   |
| ------ | ---------------------------------- | ------------ | ----------------------------- |
| `POST` | `/api/v1/roles`                    | Bearer token | Create role                   |
| `GET`  | `/api/v1/roles`                    | Bearer token | List roles for your school    |
| `GET`  | `/api/v1/roles/:id`               | Public       | Get role by ID                |
| `POST` | `/api/v1/permissions`              | Bearer token | Create permission             |
| `GET`  | `/api/v1/permissions`              | Bearer token | List all permissions          |
| `POST` | `/api/v1/roles/assign-permission`  | Bearer token | Assign permission to a role   |
| `GET`  | `/api/v1/roles/:id/permissions`    | Bearer token | Get permissions for a role    |

---

## 🧪 Test the Full Flow (Copy-Paste Ready)

### 1. Register a School

```bash
curl -s -X POST http://localhost:8081/auth/register-school \
  -H "Content-Type: application/json" \
  -d '{
    "school_name": "Springfield Elementary",
    "school_address": "123 Main St",
    "school_phone": "555-0100",
    "school_email": "info@springfield.edu",
    "admin_name": "John Doe",
    "admin_email": "john@springfield.edu",
    "admin_password": "secret123"
  }'
```

### 2. Login (save the token)

```bash
TOKEN=$(curl -s -X POST http://localhost:8081/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@springfield.edu","password":"secret123"}' \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])")

echo $TOKEN
```

> All commands below use `$TOKEN`. Run step 2 first in the same terminal.

### 3. Get My Profile

```bash
curl -s http://localhost:8081/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

### 4. Create a Role

```bash
curl -s -X POST http://localhost:8082/api/v1/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"teacher","description":"Teacher role"}'
```

> Copy the role `id` from the response for step 5.

### 5. Create a User (Teacher)

```bash
curl -s -X POST http://localhost:8081/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Smith",
    "email": "jane@springfield.edu",
    "password": "teacher123",
    "role_id": "<ROLE_ID_FROM_STEP_4>"
  }'
```

### 6. List All Users

```bash
curl -s http://localhost:8081/users \
  -H "Authorization: Bearer $TOKEN"
```

**Filter options:**

```bash
# Only active users
curl -s "http://localhost:8081/users?is_active=true" \
  -H "Authorization: Bearer $TOKEN"

# Search by name or email
curl -s "http://localhost:8081/users?search=jane" \
  -H "Authorization: Bearer $TOKEN"

# Filter by role + pagination
curl -s "http://localhost:8081/users?role_id=<ROLE_ID>&page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

### 7. Update a User

```bash
curl -s -X PATCH http://localhost:8081/users/<USER_ID> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Doe", "is_active": false}'
```

### 8. Delete a User

```bash
curl -s -X DELETE http://localhost:8081/users/<USER_ID> \
  -H "Authorization: Bearer $TOKEN"
```

### 9. Create a Permission & Assign to Role

```bash
# Create permission
curl -s -X POST http://localhost:8082/api/v1/permissions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"manage_students","description":"Can manage student records"}'

# Assign to role
curl -s -X POST http://localhost:8082/api/v1/roles/assign-permission \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role_id":"<ROLE_ID>","permission_id":"<PERMISSION_ID>"}'

# View role permissions
curl -s http://localhost:8082/api/v1/roles/<ROLE_ID>/permissions \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🔧 Useful Commands

```bash
# View logs for a specific service
docker compose logs auth-service -f
docker compose logs user-service -f

# View all logs
docker compose logs -f

# Restart a single service
docker compose restart auth-service

# Rebuild a single service
docker compose up --build auth-service -d

# Connect to a database directly
docker exec -it auth-db psql -U auth_user -d auth_db
docker exec -it user-db psql -U user_user -d user_db
```

---

## 📂 Project Structure

```
school-management/
├── docker-compose.yml
├── SETUP.md
├── auth-service/               → Port 8081
│   ├── Dockerfile
│   ├── cmd/main.go
│   └── internal/
│       ├── config/config.go
│       ├── model/models.go
│       ├── repository/auth_repository.go
│       ├── service/
│       │   ├── auth_service.go
│       │   └── user_management_service.go
│       ├── handler/
│       │   ├── auth_handler.go
│       │   └── user_handler.go
│       └── middleware/auth_middleware.go
└── user-service/               → Port 8082
    ├── Dockerfile
    ├── cmd/main.go
    └── internal/
        ├── config/config.go
        ├── model/models.go
        ├── repository/user_repository.go
        ├── service/user_service.go
        ├── handler/user_handler.go
        └── middleware/auth_middleware.go
```
