# Demo data & test credentials (xyz school)

This project does **not** give students their own login accounts. Students exist only in the **student service** database; parents and staff use **auth** logins below.

**Shared password for every demo login:** `DemoPass123!`

---

## Seed script (recommended)

From the repo root (Docker Compose running):

```bash
chmod +x scripts/seed_xyz_school_demo.sh
./scripts/seed_xyz_school_demo.sh
```

What it does (school **xyz school**, id `adc9def0-a8a2-4b8a-9d20-6b9e2893c00c`):

- Resets passwords for existing xyz users and adds **parent** + **staff** roles and users.
- Extends the **teacher** role with permissions needed to use the UI (academic, students, attendance, assignments, exams).
- Links **Aarav Sharma** to the demo parent; adds two students in **nursery section B**.
- Inserts sample **attendance**, **assignment + submission**, **exams + marks** (one published, one draft), and a **fee + partial payment**.

**After seeding:** users who are not `super_admin` should **log out and log in again** so their JWT includes the updated permission list.

---

## Auth logins (xyz school)

| Role        | Email                       | Password      | Use for |
|------------|-----------------------------|---------------|---------|
| Super admin | `xyzadmin@gmail.com`        | `DemoPass123!` | Full app; user/role management |
| Teacher 1  | `teacher@school.test`       | `DemoPass123!` | Assigned to **nursery + craft** (assignments for craft) |
| Teacher 2  | `teacher2@school.test`      | `DemoPass123!` | Assigned to **nursery + art**; demo attendance & art exam |
| Parent     | `parent.demo@xyzschool.test` | `DemoPass123!` | Submissions, results; **Aarav** is linked as child |
| Staff      | `finance.staff@xyzschool.test` | `DemoPass123!` | Fees, payments, dues |

---

## Students (nursery — not separate app logins)

Use these when filtering by class/section or when APIs ask for `student_id`. Log in as **parent** to exercise parent flows for **Aarav**.

| Name            | Section | Student ID (UUID)                    | Notes        |
|-----------------|---------|--------------------------------------|--------------|
| Aarav Sharma    | A       | `b0fb75eb-b10b-4890-a2c6-ab098abf0a1b` | Linked to parent demo user |
| Diya Patel      | A       | `f99bf4c3-fff6-470a-9667-cf4690df79b4` |              |
| Kabir Singh     | A       | `a51aaccc-65ca-4f1e-9304-b76a4172d41d` |              |
| Meera Iyer      | A       | `4f049f69-88ac-48a8-a504-7b990db51cec` |              |
| Rohan Khan      | A       | `922bb68b-4efb-4460-baff-2a8ca9525ede` |              |
| Sia Nursery-B1  | B       | `b0000001-0001-4001-8001-000000000001` | Seeded in demo script |
| Veer Nursery-B2 | B       | `b0000002-0002-4002-8002-000000000002` | Seeded in demo script |

**Class / section IDs (nursery):**

- Class **nursery**: `8f838d58-27ac-42d5-97e7-f1f38c8f0865`
- Section **A**: `322685a2-74f2-475f-a8d7-f0c0977fd92b`
- Section **B**: `ad9552d4-8e0e-4862-8b3a-5936008bd17e`

**Subjects:**

- **art**: `b9bdf360-25ec-4095-92b0-ee376725213d`
- **craft**: `74b68966-af3e-461a-ad51-c835832c3a1f`

---

## Seeded records you can click through (fixed IDs)

| Flow        | What was seeded | ID / hint |
|------------|-----------------|-----------|
| Attendance | 3 rows (yesterday), nursery-A | Teacher 2 marked Aarav / Diya / Kabir |
| Assignments | “Draw your family” (art) + one submission | Assignment `a0001000-0001-4001-8001-000000000001`, submission `a0001000-0002-4002-8002-000000000002` |
| Exams      | Published exam + marks; draft exam | Published exam `e0000001-0001-4001-8001-000000000001`, draft `e0000002-0002-4002-8002-000000000002` |
| Finance    | Class fee + partial payment for Aarav | Fee `f0000001-0001-4001-8001-000000000001`, payment `f0000002-0002-4002-8002-000000000002` |

**Dues:** in Finance, query dues with `student_id = b0fb75eb-b10b-4890-a2c6-ab098abf0a1b` to see balance after the partial payment.

---

## Flow checklist (quick)

1. **Auth** — log in as any user above; open Dashboard / profile.  
2. **Users / roles** — super admin only (`xyzadmin@gmail.com`).  
3. **Academic** — classes/sections/subjects already exist; teacher 1 has craft, teacher 2 has art.  
4. **Students** — list nursery-A / nursery-B.  
5. **Teacher assign** — super admin; extra row: teacher1 + craft.  
6. **Attendance** — teacher2, nursery, date = yesterday (or browse History).  
7. **Assignments** — teacher2 created assignment; parent submitted for Aarav.  
8. **Exams** — published results visible to parent; draft exam for publish flow as teacher2.  
9. **Finance** — staff user for fees/payments/dues.

---

## Other school in your DB (Springfield)

There is a second school (**Springfield Elementary**) with users like `john@springfield.edu`. Their passwords were **not** changed by this seed script. Use **xyz school** accounts above for a consistent `DemoPass123!` experience.
