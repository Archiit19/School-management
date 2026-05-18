# School Management System — Local Setup Guide

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (running)
- [Node.js](https://nodejs.org/) 18+ (for the web UI)
- [curl](https://curl.se/) (optional, for API testing)

---

## Access on the web (quick start)

Run these from the project root (`school-management/`).

### 1. Start all backend services

```bash
docker compose up --build -d
```

Wait until containers are healthy (about 30–60 seconds on first build):

```bash
docker compose ps
```

### 2. Install frontend dependencies (first time only)

```bash
cd frontend
npm install
cd ..
```

### 3. Start the web UI

```bash
cd frontend
npm run dev
```

Vite prints the local URL, usually:

| What | URL |
|------|-----|
| **Web app (login & UI)** | http://localhost:5173 |
| Alternate if 5173 is busy | http://localhost:5174 |

Open that URL in your browser. You should see the **login page**.

### 4. Log in

Use an account that already exists for your school, for example (see `CREDENTIALS.local.md` if you have one):

| Role | Example email | Example password |
|------|----------------|------------------|
| Admin | `school1admin@school1.com` | `123456` |
| Teacher | `school1teacher1@gmail.com` | `teacher1` |
| Student | email set at admission (e.g. `student1@school.com`) | password set in “Pupil Login” on admit |

**Student login:** the pupil must have been admitted with **Login Email** and **Initial Password** on the Students page. The email must match exactly what was saved.

After login:

- **Admin / teacher / staff** — sidebar shows Dashboard, Users, Academic, Students, Attendance, etc. (depends on role permissions).
- **Student** — sidebar shows **Dashboard** and **My Portal** (`/me`) for profile, attendance, results, assignments, and dues.

### 5. Stop when done

```bash
# Stop backend (from project root)
docker compose down

# Stop frontend: Ctrl+C in the terminal running npm run dev
```

---

## What runs where

### Backend (Docker)

| Container | Purpose | Host port |
|-----------|---------|-----------|
| `auth-db` | PostgreSQL (auth) | 5433 |
| `user-db` | PostgreSQL (roles) | 5434 |
| `academic-db` | PostgreSQL (academic) | 5435 |
| `student-db` | PostgreSQL (students) | 5436 |
| `attendance-db` | PostgreSQL (attendance) | 5437 |
| `exam-db` | PostgreSQL (exams) | 5438 |
| `finance-db` | PostgreSQL (finance) | 5439 |
| `auth-service` | Login, users, JWT | **8081** |
| `user-service` | Roles & permissions | **8082** |
| `academic-service` | Classes, assignments | **8083** |
| `student-service` | Student admission | **8084** |
| `attendance-service` | Attendance | **8085** |
| `exam-service` | Exams & results | **8086** |
| `finance-service` | Fees & dues | **8087** |

The frontend does **not** run in Docker. It proxies API calls to these ports (see `frontend/vite.config.js`).

### Frontend → backend mapping (browser)

The UI calls paths like `/api/auth/...`; Vite forwards them to localhost:

| Browser path prefix | Backend |
|---------------------|---------|
| `/api/auth` | http://localhost:8081 |
| `/api/users` | http://localhost:8082 |
| `/api/academic` | http://localhost:8083 |
| `/api/students` | http://localhost:8084 |
| `/api/attendance` | http://localhost:8085 |
| `/api/exams` | http://localhost:8086 |
| `/api/finance` | http://localhost:8087 |

**Login flow in the UI:** `POST /api/auth/auth/login` → then `GET /api/auth/auth/me` (see `frontend/src/api/client.js` and `frontend/src/context/AuthContext.jsx`).

---

## Verify backend is up

```bash
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health
curl http://localhost:8085/health
curl http://localhost:8086/health
curl http://localhost:8087/health
```

Each should return a JSON status like `{"status":"...-service is running"}`.

---

## Swagger UI (API docs in browser)

With Docker running, open:

| Service | Swagger URL |
|---------|-------------|
| Auth | http://localhost:8081/swagger/index.html |
| User | http://localhost:8082/swagger/index.html |
| Academic | http://localhost:8083/swagger/index.html |
| Student | http://localhost:8084/swagger/index.html |
| Attendance | http://localhost:8085/swagger/index.html |
| Exam | http://localhost:8086/swagger/index.html |
| Finance | http://localhost:8087/swagger/index.html |

---

## Rebuild after code changes

```bash
# Rebuild all services
docker compose up --build -d

# Rebuild one service only (example)
docker compose up --build auth-service -d
```

Frontend picks up changes automatically while `npm run dev` is running (hot reload). Restart dev server if you change `vite.config.js`.

---

## Stop / reset

```bash
# Stop containers (data preserved)
docker compose down

# Stop AND wipe all database data (fresh start)
docker compose down -v
```

---

## First-time school (no account yet)

If you have no school in the database, register via API (or implement register in UI later):

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

Then log in on the web app with `john@springfield.edu` / `secret123`.

Template roles (`super_admin`, `teacher`, `parent`, `staff`, `student`) are created automatically for the new school.

---

## Admit a student with login (web)

1. Log in as **admin**.
2. Open **Academic Structure** — create at least one class.
3. Open **Students** → **Admit Student**.
4. Fill name, class, and under **Pupil Login (optional)** enter **Login Email** and **Initial Password** (both required to create a login).
5. Pupil logs out; logs in with that email and password → opens **My Portal**.

---

## Useful commands

```bash
# View logs
docker compose logs -f
docker compose logs auth-service -f
docker compose logs student-service -f

# Restart one service
docker compose restart auth-service

# Connect to a database
docker exec -it auth-db psql -U auth_user -d auth_db
docker exec -it student-db psql -U student_user -d student_db
```

---

## More documentation

- **Flows & all APIs:** `FLOWS_AND_APIS.md`
- **Local credentials (if present):** `CREDENTIALS.local.md` (not committed; create your own)

---

## Project structure (overview)

```
school-management/
├── docker-compose.yml      # All backends + databases
├── SETUP.md                # This file
├── FLOWS_AND_APIS.md
├── frontend/               # React + Vite (npm run dev → :5173)
│   ├── src/
│   │   ├── api/client.js   # API client + proxy paths
│   │   ├── pages/          # Dashboard, Students, MyPortal, …
│   │   └── components/Layout.jsx
│   └── vite.config.js
├── auth-service/           # :8081
├── user-service/           # :8082
├── academic-service/       # :8083
├── student-service/        # :8084
├── attendance-service/     # :8085
├── exam-service/           # :8086
└── finance-service/        # :8087
```
