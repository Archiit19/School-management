#!/usr/bin/env bash
# Seeds the "xyz school" tenant (school_id adc9def0-a8a2-4b8a-9d20-6b9e2893c00c) with demo users,
# role permissions, students, attendance, assignment + submission, exams + marks, and a fee + payment.
#
# Prerequisites: docker compose stack running (same DBs as docker-compose.yml).
# Password for all demo auth users set here: DemoPass123!
#
# After running: have each non–super-admin user log out and log in again so their JWT picks up new permissions.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DEMO_PW_HASH='$2a$10$9mW815OyUdxdILWDNsBDKu55z2cIPG5LukjxF00Q2ECHC5DGOlB.i'

SCHOOL='adc9def0-a8a2-4b8a-9d20-6b9e2893c00c'
CLASS='8f838d58-27ac-42d5-97e7-f1f38c8f0865'
SEC_A='322685a2-74f2-475f-a8d7-f0c0977fd92b'
SEC_B='ad9552d4-8e0e-4862-8b3a-5936008bd17e'
SUBJ_ART='b9bdf360-25ec-4095-92b0-ee376725213d'
SUBJ_CRAFT='74b68966-af3e-461a-ad51-c835832c3a1f'

SUPER_USER='95e683be-6f25-444e-9731-74a885ea5b16'
TEACHER1='21ea3636-5403-43eb-bcca-05a6da9fa579'
TEACHER2='82996dbb-814c-49f4-a460-b6d911e2bae7'
TEACHER_ROLE='55bb6173-6eb0-4dd6-a7b2-4cbf56d6df8d'

PARENT_ROLE='11111111-1111-4111-8111-111111111101'
STAFF_ROLE='11111111-1111-4111-8111-111111111102'
PARENT_USER='22222222-2222-4222-8222-222222222201'
STAFF_USER='22222222-2222-4222-8222-222222222202'

STU_AARAV='b0fb75eb-b10b-4890-a2c6-ab098abf0a1b'
STU_DIYA='f99bf4c3-fff6-470a-9667-cf4690df79b4'
STU_KABIR='a51aaccc-65ca-4f1e-9304-b76a4172d41d'

STU_B1='b0000001-0001-4001-8001-000000000001'
STU_B2='b0000002-0002-4002-8002-000000000002'

ASSIGN_ID='a0001000-0001-4001-8001-000000000001'
SUBMIT_ID='a0001000-0002-4002-8002-000000000002'

EXAM_PUB='e0000001-0001-4001-8001-000000000001'
EXAM_DRAFT='e0000002-0002-4002-8002-000000000002'

FEE_ID='f0000001-0001-4001-8001-000000000001'
PAY_ID='f0000002-0002-4002-8002-000000000002'

ATT1='d0000001-0001-4001-8001-000000000001'
ATT2='d0000002-0002-4002-8002-000000000002'
ATT3='d0000003-0003-4003-8003-000000000003'

echo "==> user-db: parent + staff roles"
docker exec -i user-db psql -U user_user -d user_db <<SQL
\set ON_ERROR_STOP on
INSERT INTO roles (id, school_id, name, description, created_at, updated_at)
VALUES
  ('$PARENT_ROLE'::uuid, '$SCHOOL'::uuid, 'parent', 'Demo parent role', now(), now()),
  ('$STAFF_ROLE'::uuid, '$SCHOOL'::uuid, 'staff', 'Demo finance/office staff', now(), now())
ON CONFLICT (id) DO NOTHING;

-- Teacher: full school-day teaching + academic flows (no user/role admin, no fees)
INSERT INTO role_permissions (id, role_id, permission_id, created_at)
SELECT gen_random_uuid(), '$TEACHER_ROLE'::uuid, p.id, now()
FROM permissions p
WHERE p.name IN (
  'view_academic',
  'mark_attendance', 'view_attendance',
  'view_students', 'update_student',
  'create_assignment', 'view_assignments', 'submit_assignment',
  'create_exam', 'enter_marks', 'publish_results', 'view_results'
)
AND NOT EXISTS (
  SELECT 1 FROM role_permissions rp
  WHERE rp.role_id = '$TEACHER_ROLE'::uuid AND rp.permission_id = p.id
);

INSERT INTO role_permissions (id, role_id, permission_id, created_at)
SELECT gen_random_uuid(), '$PARENT_ROLE'::uuid, p.id, now()
FROM permissions p
WHERE p.name IN ('view_assignments', 'submit_assignment', 'view_results', 'view_students')
AND NOT EXISTS (
  SELECT 1 FROM role_permissions rp
  WHERE rp.role_id = '$PARENT_ROLE'::uuid AND rp.permission_id = p.id
);

INSERT INTO role_permissions (id, role_id, permission_id, created_at)
SELECT gen_random_uuid(), '$STAFF_ROLE'::uuid, p.id, now()
FROM permissions p
WHERE p.name IN ('create_fee', 'record_payment', 'view_dues', 'view_students')
AND NOT EXISTS (
  SELECT 1 FROM role_permissions rp
  WHERE rp.role_id = '$STAFF_ROLE'::uuid AND rp.permission_id = p.id
);
SQL

echo "==> auth-db: reset passwords + parent/staff users"
docker exec -i auth-db psql -U auth_user -d auth_db <<SQL
\set ON_ERROR_STOP on
UPDATE users SET password = '$DEMO_PW_HASH', updated_at = now()
WHERE email IN ('xyzadmin@gmail.com', 'teacher@school.test', 'teacher2@school.test');

INSERT INTO users (id, school_id, name, email, password, role_id, is_active, created_at, updated_at)
VALUES
  ('$PARENT_USER'::uuid, '$SCHOOL'::uuid, 'Demo Parent', 'parent.demo@xyzschool.test', '$DEMO_PW_HASH', '$PARENT_ROLE'::uuid, true, now(), now()),
  ('$STAFF_USER'::uuid, '$SCHOOL'::uuid, 'Demo Finance Staff', 'finance.staff@xyzschool.test', '$DEMO_PW_HASH', '$STAFF_ROLE'::uuid, true, now(), now())
ON CONFLICT (email) DO UPDATE SET
  password = EXCLUDED.password,
  role_id = EXCLUDED.role_id,
  name = EXCLUDED.name,
  is_active = true,
  updated_at = now();
SQL

echo "==> academic-db: teacher1 → nursery + craft (second assignment scenario)"
docker exec -i academic-db psql -U academic_user -d academic_db <<SQL
\set ON_ERROR_STOP on
INSERT INTO teacher_assignments (id, school_id, teacher_user_id, class_id, subject_id, created_at, updated_at)
SELECT gen_random_uuid(), '$SCHOOL'::uuid, '$TEACHER1'::uuid, '$CLASS'::uuid, '$SUBJ_CRAFT'::uuid, now(), now()
WHERE NOT EXISTS (
  SELECT 1 FROM teacher_assignments ta
  WHERE ta.school_id = '$SCHOOL'::uuid
    AND ta.teacher_user_id = '$TEACHER1'::uuid
    AND ta.class_id = '$CLASS'::uuid
    AND ta.subject_id = '$SUBJ_CRAFT'::uuid
);
SQL

echo "==> student-db: link parent + section B students"
docker exec -i student-db psql -U student_user -d student_db <<SQL
\set ON_ERROR_STOP on
UPDATE students SET parent_user_id = '$PARENT_USER'::uuid, updated_at = now()
WHERE id = '$STU_AARAV'::uuid;

INSERT INTO students (id, school_id, first_name, last_name, class_id, section_id, is_active, created_at, updated_at)
VALUES
  ('$STU_B1'::uuid, '$SCHOOL'::uuid, 'Sia', 'Nursery-B1', '$CLASS'::uuid, '$SEC_B'::uuid, true, now(), now()),
  ('$STU_B2'::uuid, '$SCHOOL'::uuid, 'Veer', 'Nursery-B2', '$CLASS'::uuid, '$SEC_B'::uuid, true, now(), now())
ON CONFLICT (id) DO NOTHING;
SQL

echo "==> attendance-db: sample attendances (yesterday)"
docker exec -i attendance-db psql -U attendance_user -d attendance_db <<SQL
\set ON_ERROR_STOP on
INSERT INTO attendances (id, school_id, teacher_user_id, student_id, class_id, section_id, subject_id, date, status, remarks, created_at, updated_at)
VALUES
  ('$ATT1'::uuid, '$SCHOOL'::uuid, '$TEACHER2'::uuid, '$STU_AARAV'::uuid, '$CLASS'::uuid, '$SEC_A'::uuid, NULL, CURRENT_DATE - 1, 'present', 'Demo seed', now(), now()),
  ('$ATT2'::uuid, '$SCHOOL'::uuid, '$TEACHER2'::uuid, '$STU_DIYA'::uuid, '$CLASS'::uuid, '$SEC_A'::uuid, NULL, CURRENT_DATE - 1, 'absent', 'Demo seed', now(), now()),
  ('$ATT3'::uuid, '$SCHOOL'::uuid, '$TEACHER2'::uuid, '$STU_KABIR'::uuid, '$CLASS'::uuid, '$SEC_A'::uuid, NULL, CURRENT_DATE - 1, 'late', 'Demo seed', now(), now())
ON CONFLICT (id) DO NOTHING;
SQL

echo "==> academic-db: assignment + parent submission"
docker exec -i academic-db psql -U academic_user -d academic_db <<SQL
\set ON_ERROR_STOP on
INSERT INTO assignments (id, school_id, teacher_user_id, class_id, subject_id, title, description, material_url, due_date, created_at, updated_at)
VALUES (
  '$ASSIGN_ID'::uuid, '$SCHOOL'::uuid, '$TEACHER2'::uuid, '$CLASS'::uuid, '$SUBJ_ART'::uuid,
  'Draw your family', 'Demo homework for Flow 7.', 'https://example.com/art-rubric',
  CURRENT_DATE + 7, now(), now()
) ON CONFLICT (id) DO NOTHING;

INSERT INTO submissions (id, school_id, assignment_id, student_id, submitted_by, content, material_url, created_at, updated_at)
VALUES (
  '$SUBMIT_ID'::uuid, '$SCHOOL'::uuid, '$ASSIGN_ID'::uuid, '$STU_AARAV'::uuid, '$PARENT_USER'::uuid,
  'Submitted drawing description (demo).', 'https://example.com/parent-upload/demo.png',
  now(), now()
) ON CONFLICT (id) DO NOTHING;
SQL

echo "==> exam-db: published + draft exams + marks"
docker exec -i exam-db psql -U exam_user -d exam_db <<SQL
\set ON_ERROR_STOP on
INSERT INTO exams (id, school_id, class_id, section_id, subject_id, title, exam_date, total_marks, is_published, created_by, created_at, updated_at)
VALUES
  ('$EXAM_PUB'::uuid, '$SCHOOL'::uuid, '$CLASS'::uuid, '$SEC_A'::uuid, '$SUBJ_ART'::uuid,
   'Nursery Art — Term check-in', CURRENT_DATE - 14, 50, true, '$TEACHER2'::uuid, now(), now()),
  ('$EXAM_DRAFT'::uuid, '$SCHOOL'::uuid, '$CLASS'::uuid, '$SEC_A'::uuid, '$SUBJ_ART'::uuid,
   'Nursery Art — Draft quiz (unpublished)', CURRENT_DATE + 3, 20, false, '$TEACHER2'::uuid, now(), now())
ON CONFLICT (id) DO NOTHING;

INSERT INTO marks (id, school_id, exam_id, student_id, marks_obtained, remarks, created_by, created_at, updated_at)
SELECT gen_random_uuid(), '$SCHOOL'::uuid, '$EXAM_PUB'::uuid, sid, mk, 'Demo seed', '$TEACHER2'::uuid, now(), now()
FROM (VALUES
  ('$STU_AARAV'::uuid, 44::numeric),
  ('$STU_DIYA'::uuid, 38::numeric),
  ('$STU_KABIR'::uuid, 47::numeric)
) AS v(sid, mk)
WHERE NOT EXISTS (SELECT 1 FROM marks m WHERE m.exam_id = '$EXAM_PUB'::uuid AND m.student_id = v.sid);
SQL

echo "==> finance-db: class fee + partial payment (Aarav)"
docker exec -i finance-db psql -U finance_user -d finance_db <<SQL
\set ON_ERROR_STOP on
INSERT INTO fees (id, school_id, title, description, amount, class_id, section_id, student_id, due_date, is_active, created_by, created_at, updated_at)
VALUES (
  '$FEE_ID'::uuid, '$SCHOOL'::uuid, 'Nursery annual supplies', 'Demo class-level fee', 2500, '$CLASS'::uuid, NULL, NULL,
  CURRENT_DATE + 30, true, '$SUPER_USER'::uuid, now(), now()
) ON CONFLICT (id) DO NOTHING;

INSERT INTO payments (id, school_id, fee_id, student_id, amount_paid, payment_date, method, reference, received_by, created_at, updated_at)
VALUES (
  '$PAY_ID'::uuid, '$SCHOOL'::uuid, '$FEE_ID'::uuid, '$STU_AARAV'::uuid, 800, CURRENT_DATE - 2,
  'cash', 'DEMO-SEED-1', '$STAFF_USER'::uuid, now(), now()
)
ON CONFLICT (id) DO NOTHING;
SQL

echo ""
echo "Done. See DEMO_CREDENTIALS.md for logins and entity IDs."
echo "Important: teachers/parent/staff must re-login to refresh JWT permissions."
