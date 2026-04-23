# School Management — Flows & APIs

Reference for **business flows** and the **HTTP APIs** that implement them.  
Default local ports (from `docker-compose.yml`): auth `8081`, user `8082`, academic `8083`, student `8084`, attendance `8085`, exam `8086`, finance `8087`.

Legend: **Public** = no JWT. **Bearer** = `Authorization: Bearer <token>`.

---

## Flow 1 — Tenant & identity (Auth Service)

**Goal:** Create the school and first admin, authenticate, read session.

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| `POST` | `http://localhost:8081/auth/register-school` | Public | Creates school + `super_admin` role (via user-service) + admin user; returns JWT |
| `POST` | `http://localhost:8081/auth/login` | Public | Returns JWT |
| `GET` | `http://localhost:8081/auth/me` | Bearer | Current user profile |
| `GET` | `http://localhost:8081/health` | Public | Liveness |
| `GET` | `http://localhost:8081/swagger/*` | Public | Swagger UI |

---

## Flow 2 — Roles & permissions (User Service)

**Goal:** Define roles (teacher, parent, staff, …), optional permissions, map permissions to roles.

**Base path:** `http://localhost:8082/api/v1`

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| `POST` | `/roles/internal` | Public | Used by auth-service during school registration |
| `GET` | `/roles/:id` | Public | Role by UUID (used by auth to resolve role name) |
| `POST` | `/roles` | Bearer | Create role for your school |
| `GET` | `/roles` | Bearer | List roles for your school |
| `POST` | `/permissions` | Bearer | Create global permission |
| `GET` | `/permissions` | Bearer | List all permissions |
| `POST` | `/roles/assign-permission` | Bearer | Link permission → role |
| `GET` | `/roles/:id/permissions` | Bearer | List permissions for a role |
| `GET` | `http://localhost:8082/health` | Public | Liveness |
| `GET` | `http://localhost:8082/swagger/*` | Public | Swagger UI |

---

## Flow 3 — School structure (Academic Service)

**Goal:** Classes, sections, subjects; view full tree.

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| `POST` | `http://localhost:8083/classes` | Bearer | Create class |
| `POST` | `http://localhost:8083/sections` | Bearer | Create section under class |
| `POST` | `http://localhost:8083/subjects` | Bearer | Create subject (optional section) |
| `GET` | `http://localhost:8083/classes` | Bearer | List classes with sections & subjects |
| `GET` | `http://localhost:8083/health` | Public | Liveness |
| `GET` | `http://localhost:8083/swagger/*` | Public | Swagger UI |

---

## Flow 4 — Student admission (Student Service)

**Goal:** Enroll students; assign class/section; optional parent link (auth user with `parent` role).

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| `POST` | `http://localhost:8084/students` | Bearer | Create student |
| `GET` | `http://localhost:8084/students` | Bearer | List (query: page, limit, search, class_id, section_id, parent_user_id, is_active) |
| `PATCH` | `http://localhost:8084/students/:id` | Bearer | Update student |
| `GET` | `http://localhost:8084/health` | Public | Liveness |
| `GET` | `http://localhost:8084/swagger/*` | Public | Swagger UI |

---

## Flow 5 — Teacher assignment (Academic Service)

**Goal:** Map a teacher user to a class + subject (required before teacher-created assignments for that class/subject).

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| `POST` | `http://localhost:8083/teacher-assignments` | Bearer | Assign teacher → class + subject |
| `GET` | `http://localhost:8083/teacher-assignments` | Bearer | List (query: teacher_user_id, class_id, subject_id) |

---

## Flow 6 — Attendance (Attendance Service)

**Goal:** Mark, list, and correct daily attendance.

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| `POST` | `http://localhost:8085/attendance` | Bearer | Mark attendance (role: teacher or super_admin) |
| `GET` | `http://localhost:8085/attendance` | Bearer | List with filters (page, limit, date, student_id, class_id, section_id, subject_id, status) |
| `PATCH` | `http://localhost:8085/attendance/:id` | Bearer | Edit (teacher: own rows; super_admin: any) |
| `GET` | `http://localhost:8085/health` | Public | Liveness |
| `GET` | `http://localhost:8085/swagger/*` | Public | Swagger UI |

---

## Flow 7 — Assignments & study material (Academic Service)

**Goal:** Teacher publishes work; student submits (optional material URLs).

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| `POST` | `http://localhost:8083/assignments` | Bearer | Create assignment (teacher needs prior teacher-assignment for class+subject) |
| `GET` | `http://localhost:8083/assignments` | Bearer | List (query: class_id, subject_id, teacher_id) |
| `POST` | `http://localhost:8083/submissions` | Bearer | Submit for assignment + student |

---

## Flow 8 — Exams & results (Exam Service)

**Goal:** Exam lifecycle — create exam, enter marks, publish, view results.

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| `POST` | `http://localhost:8086/exams` | Bearer | Create exam |
| `POST` | `http://localhost:8086/marks` | Bearer | Enter / upsert marks |
| `POST` | `http://localhost:8086/results/publish` | Bearer | Publish results for an exam |
| `GET` | `http://localhost:8086/results` | Bearer | List computed results (query: exam_id, student_id, class_id) |
| `GET` | `http://localhost:8086/health` | Public | Liveness |
| `GET` | `http://localhost:8086/swagger/*` | Public | Swagger UI |

---

## Flow 9 — Fees & payments (Finance Service)

**Goal:** Fee structure, record payments, view dues.

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| `POST` | `http://localhost:8087/fees` | Bearer | Create fee (super_admin or staff) |
| `POST` | `http://localhost:8087/payments` | Bearer | Record payment (super_admin or staff) |
| `GET` | `http://localhost:8087/dues` | Bearer | Dues (query: student_id, class_id, section_id) |
| `GET` | `http://localhost:8087/health` | Public | Liveness |
| `GET` | `http://localhost:8087/swagger/*` | Public | Swagger UI |

---

## Cross-cutting — User management (Auth Service)

**Goal:** Super admin creates staff / teachers / parents (needs `role_id` from user-service).

| Method | Path | Auth | Notes |
|--------|------|------|--------|
| `POST` | `http://localhost:8081/users` | Bearer + **super_admin** | Create user |
| `GET` | `http://localhost:8081/users` | Bearer + **super_admin** | List users |
| `GET` | `http://localhost:8081/users/:id` | Bearer + **super_admin** | Get user |
| `PATCH` | `http://localhost:8081/users/:id` | Bearer + **super_admin** | Update user |
| `DELETE` | `http://localhost:8081/users/:id` | Bearer + **super_admin** | Delete user |

---

## OpenAPI (Swagger) contract files per service

| Service | `swagger.yaml` |
|---------|----------------|
| auth-service | `auth-service/docs/swagger.yaml` |
| user-service | `user-service/docs/swagger.yaml` |
| academic-service | `academic-service/docs/swagger.yaml` |
| student-service | `student-service/docs/swagger.yaml` |
| attendance-service | `attendance-service/docs/swagger.yaml` |
| exam-service | `exam-service/docs/swagger.yaml` |
| finance-service | `finance-service/docs/swagger.yaml` |

Each service also has `docs/swagger.json` and `docs/docs.go`.

---

## Suggested bootstrap order

1. Flow 1 — register school / login  
2. Flow 2 — roles (and optional permissions)  
3. Flow 3 — classes → sections → subjects  
4. Cross-cutting — create users (parent, teacher, staff) via `/users`  
5. Flow 5 — teacher-assignments  
6. Flow 4 — students  
7. Flow 6 — attendance  
8. Flow 7 — assignments → submissions  
9. Flow 8 — exams → marks → publish → results  
10. Flow 9 — fees → payments → dues  

---

*This file is the canonical flow/API index; `SETUP.md` may list fewer containers—use `docker compose ps` for your live stack.*
