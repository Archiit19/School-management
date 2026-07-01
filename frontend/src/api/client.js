const SERVICES = {
  auth: "/api/auth",
  users: "/api/users",
  schools: "/api/schools",
  academic: "/api/academic",
  students: "/api/students",
  attendance: "/api/attendance",
  exams: "/api/exams",
  finance: "/api/finance",
};

function getToken() {
  return localStorage.getItem("token") || "";
}

function pupilQuery(query, studentId) {
  const q = { ...(query || {}) };
  if (studentId) q.student_id = studentId;
  return q;
}

export class ApiError extends Error {
  constructor(message, status, data) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.data = data;
  }
}

function mapStatusToMessage(status, fallback) {
  if (fallback) return fallback;
  if (status === 403) return "Forbidden: you do not have permission for this action.";
  if (status === 409) return "Conflict: the action conflicts with existing data.";
  if (status === 503) return "Service unavailable: please try again in a moment.";
  return `Request failed (${status})`;
}

async function request(service, path, { method = "GET", body, query } = {}) {
  const base = SERVICES[service];
  if (!base) throw new Error(`Unknown service: ${service}`);

  let url = `${base}${path}`;
  if (query) {
    const params = new URLSearchParams();
    Object.entries(query).forEach(([k, v]) => {
      if (v !== undefined && v !== null && v !== "") params.append(k, v);
    });
    const qs = params.toString();
    if (qs) url += `?${qs}`;
  }

  const headers = { "Content-Type": "application/json" };
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(url, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  const data = await res.json().catch(() => null);

  if (!res.ok) {
    const msg = mapStatusToMessage(res.status, data?.error);
    throw new ApiError(msg, res.status, data);
  }

  return data;
}

/** Flatten role_data onto user records for pages that expect legacy student fields. */
function normalizeStudentUser(user) {
  const roleData = user.role_data || {};
  const fullName = (user.name || "").trim();
  const parts = fullName.split(/\s+/).filter(Boolean);
  return {
    ...user,
    class_id: roleData.class_id ?? user.class_id,
    section_id: roleData.section_id ?? user.section_id,
    first_name: roleData.first_name || parts[0] || fullName || "—",
    last_name: roleData.last_name || parts.slice(1).join(" ") || "",
    student_code: roleData.student_code || user.student_code || "",
    parent_name: roleData.parent_name || "",
    contact_number: roleData.contact_number || "",
  };
}

/** Map user-service pupil records to the shape attendance/exams pages expect. */
function normalizeStudentRecord(user, enrollments) {
  const enrollment = enrollments.find((e) => String(e.user_id) === String(user.id));
  const merged = {
    ...user,
    role_data: { ...(user.role_data || {}) },
  };
  if (enrollment) {
    merged.role_data.class_id = enrollment.class_id;
    if (enrollment.section_id) merged.role_data.section_id = enrollment.section_id;
  }
  return normalizeStudentUser(merged);
}

export const authApi = {
  signup: (body) => request("auth", "/auth/signup", { method: "POST", body }),
  registerSchool: (body) => request("auth", "/auth/register-school", { method: "POST", body }),
  login: (body) => request("auth", "/auth/login", { method: "POST", body }),
  me: () => request("auth", "/auth/me"),
  selectSchool: (schoolId) => request("auth", "/auth/select-school", { method: "POST", body: { school_id: schoolId } }),
  exitSchool: () => request("auth", "/auth/exit-school", { method: "POST" }),
  health: () => request("auth", "/health"),
};

export const userMgmtApi = {
  create: (body) => request("users", "/users", { method: "POST", body }),
  list: (query) => request("users", "/users", { query }),
  getById: (id) => request("users", `/users/${id}`),
  update: (id, body) => request("users", `/users/${id}`, { method: "PATCH", body }),
  remove: (id) => request("users", `/users/${id}`, { method: "DELETE" }),
  health: () => request("users", "/health"),
};

export const rolesApi = {
  create: (body) => request("auth", "/api/v1/roles", { method: "POST", body }),
  list: () => request("auth", "/api/v1/roles"),
  getById: (id) => request("auth", `/api/v1/roles/${id}`),
  getFields: (id) => request("auth", `/api/v1/roles/${id}/fields`),
  updateFields: (id, fields) => request("auth", `/api/v1/roles/${id}/fields`, { method: "PUT", body: { fields } }),
  health: () => request("auth", "/health"),
};

export const permissionsApi = {
  create: (body) => request("auth", "/api/v1/permissions", { method: "POST", body }),
  list: () => request("auth", "/api/v1/permissions"),
  assign: (body) => request("auth", "/api/v1/roles/assign-permission", { method: "POST", body }),
  forRole: (roleId) => request("auth", `/api/v1/roles/${roleId}/permissions`),
  removeFromRole: (roleId, permissionId) =>
    request("auth", `/api/v1/roles/${roleId}/permissions/${permissionId}`, { method: "DELETE" }),
};

export const academicApi = {
  createClass: (body) => request("academic", "/classes", { method: "POST", body }),
  createSection: (body) => request("academic", "/sections", { method: "POST", body }),
  createSubject: (body) => request("academic", "/subjects", { method: "POST", body }),
  getClasses: () => request("academic", "/classes"),
  getEnrollments: (query) => request("academic", "/enrollments", { query }),
  createTeacherAssignment: (body) => request("academic", "/teacher-assignments", { method: "POST", body }),
  getTeacherAssignments: (query) => request("academic", "/teacher-assignments", { query }),
  updateTeacherAssignment: (id, body) => request("academic", `/teacher-assignments/${id}`, { method: "PATCH", body }),
  deleteTeacherAssignment: (id) => request("academic", `/teacher-assignments/${id}`, { method: "DELETE" }),
  createAssignment: (body) => request("academic", "/assignments", { method: "POST", body }),
  getAssignments: (query) => request("academic", "/assignments", { query }),
  getAssignmentSubmissions: (assignmentId) =>
    request("academic", `/assignments/${assignmentId}/submissions`),
  reviewSubmission: (id, body) =>
    request("academic", `/submissions/${id}`, { method: "PATCH", body }),
  createSubmission: (body) => request("academic", "/submissions", { method: "POST", body }),
  getMyAssignments: (studentId) =>
    request("academic", "/assignments/me", { query: pupilQuery({}, studentId) }),
  getMySubmissions: (studentId) =>
    request("academic", "/submissions/me", { query: pupilQuery({}, studentId) }),
  submitMine: (body) => request("academic", "/submissions/me", { method: "POST", body }),
  getMyAcademic: (studentId) =>
    request("academic", "/academic/me", { query: pupilQuery({}, studentId) }),
  getMyEnrollment: (studentId) =>
    request("academic", "/enrollments/me", { query: pupilQuery({}, studentId) }),
  health: () => request("academic", "/health"),
};

export const studentApi = {
  async list(query) {
    const enrollRes = await academicApi.getEnrollments({
      class_id: query.class_id,
      section_id: query.section_id || undefined,
    });
    const enrollments = enrollRes.enrollments || [];
    if (enrollments.length === 0) {
      return { students: [], users: [], total: 0 };
    }
    const ids = enrollments.map((e) => e.user_id).join(",");
    const res = await request("users", "/users", {
      query: { ids, limit: query.limit || 200, page: 1 },
    });
    const users = (res.users || []).map((u) => normalizeStudentRecord(u, enrollments));
    return { students: users, users, total: res.total || users.length };
  },
  create: (body) => request("users", "/users", { method: "POST", body }),
  update: (id, body) => request("users", `/users/${id}`, { method: "PATCH", body }),
  getMe: async () => normalizeStudentUser(await request("users", "/users/me")),
  getMyChildren: () => request("users", "/users/me/children"),
  getChild: (id) => request("users", `/users/me/children/${id}`),
  health: () => request("users", "/health"),
};

export const attendanceApi = {
  create: (body) => request("attendance", "/attendance", { method: "POST", body }),
  bulkCreate: (body) => request("attendance", "/attendance/bulk", { method: "POST", body }),
  list: (query) => request("attendance", "/attendance", { query }),
  update: (id, body) => request("attendance", `/attendance/${id}`, { method: "PATCH", body }),
  stats: (query) => request("attendance", "/attendance/stats", { query }),
  createTeacher: (body) =>
    request("attendance", "/teacher-attendance", { method: "POST", body }),
  bulkCreateTeacher: (body) =>
    request("attendance", "/teacher-attendance/bulk", { method: "POST", body }),
  listTeacher: (query) => request("attendance", "/teacher-attendance", { query }),
  updateTeacher: (id, body) =>
    request("attendance", `/teacher-attendance/${id}`, { method: "PATCH", body }),
  statsTeacher: (query) => request("attendance", "/teacher-attendance/stats", { query }),
  getMine: (query, studentId) =>
    request("attendance", "/attendance/me", { query: pupilQuery(query, studentId) }),
  myStats: (query, studentId) =>
    request("attendance", "/attendance/me/stats", { query: pupilQuery(query, studentId) }),
  health: () => request("attendance", "/health"),
};

export const examApi = {
  createExam: (body) => request("exams", "/exams", { method: "POST", body }),
  updateExam: (id, body) => request("exams", `/exams/${id}`, { method: "PATCH", body }),
  completeExam: (id) => request("exams", `/exams/${id}/complete`, { method: "POST" }),
  deleteExam: (id) => request("exams", `/exams/${id}`, { method: "DELETE" }),
  getExams: (query) => request("exams", "/exams", { query }),
  getMyExams: (query, studentId) =>
    request("exams", "/exams/me", { query: pupilQuery(query, studentId) }),
  enterMarks: (body) => request("exams", "/marks", { method: "POST", body }),
  publish: (body) => request("exams", "/results/publish", { method: "POST", body }),
  getResults: (query) => request("exams", "/results", { query }),
  getMyResults: (query, studentId) =>
    request("exams", "/results/me", { query: pupilQuery(query, studentId) }),
  health: () => request("exams", "/health"),
};

export const schoolApi = {
  create: (body) => request("schools", "/schools", { method: "POST", body }),
  listMine: () => request("schools", "/schools/mine"),
  list: (query) => request("schools", "/schools", { query }),
  getMe: () => request("schools", "/schools/me"),
  updateMe: (body) => request("schools", "/schools/me", { method: "PATCH", body }),
  getById: (id) => request("schools", `/schools/${id}`),
  update: (id, body) => request("schools", `/schools/${id}`, { method: "PATCH", body }),
  remove: (id) => request("schools", `/schools/${id}`, { method: "DELETE" }),
  health: () => request("schools", "/health"),
};

export const financeApi = {
  createFee: (body) => request("finance", "/fees", { method: "POST", body }),
  recordPayment: (body) => request("finance", "/payments", { method: "POST", body }),
  getDues: (query) => request("finance", "/dues", { query }),
  getMyDues: (studentId) => request("finance", "/dues/me", { query: pupilQuery({}, studentId) }),
  health: () => request("finance", "/health"),
};
