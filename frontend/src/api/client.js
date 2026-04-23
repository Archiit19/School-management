const SERVICES = {
  auth: "/api/auth",
  users: "/api/users",
  academic: "/api/academic",
  students: "/api/students",
  attendance: "/api/attendance",
  exams: "/api/exams",
  finance: "/api/finance",
};

function getToken() {
  return localStorage.getItem("token") || "";
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
    const msg = data?.error || `Request failed (${res.status})`;
    throw new Error(msg);
  }

  return data;
}

export const authApi = {
  registerSchool: (body) => request("auth", "/auth/register-school", { method: "POST", body }),
  login: (body) => request("auth", "/auth/login", { method: "POST", body }),
  me: () => request("auth", "/auth/me"),
  health: () => request("auth", "/health"),
};

export const userMgmtApi = {
  create: (body) => request("auth", "/users", { method: "POST", body }),
  list: (query) => request("auth", "/users", { query }),
  getById: (id) => request("auth", `/users/${id}`),
  update: (id, body) => request("auth", `/users/${id}`, { method: "PATCH", body }),
  remove: (id) => request("auth", `/users/${id}`, { method: "DELETE" }),
};

export const rolesApi = {
  create: (body) => request("users", "/api/v1/roles", { method: "POST", body }),
  list: () => request("users", "/api/v1/roles"),
  getById: (id) => request("users", `/api/v1/roles/${id}`),
  health: () => request("users", "/health"),
};

export const permissionsApi = {
  create: (body) => request("users", "/api/v1/permissions", { method: "POST", body }),
  list: () => request("users", "/api/v1/permissions"),
  assign: (body) => request("users", "/api/v1/roles/assign-permission", { method: "POST", body }),
  forRole: (roleId) => request("users", `/api/v1/roles/${roleId}/permissions`),
};

export const academicApi = {
  createClass: (body) => request("academic", "/classes", { method: "POST", body }),
  createSection: (body) => request("academic", "/sections", { method: "POST", body }),
  createSubject: (body) => request("academic", "/subjects", { method: "POST", body }),
  getClasses: () => request("academic", "/classes"),
  createTeacherAssignment: (body) => request("academic", "/teacher-assignments", { method: "POST", body }),
  getTeacherAssignments: (query) => request("academic", "/teacher-assignments", { query }),
  createAssignment: (body) => request("academic", "/assignments", { method: "POST", body }),
  getAssignments: (query) => request("academic", "/assignments", { query }),
  createSubmission: (body) => request("academic", "/submissions", { method: "POST", body }),
  health: () => request("academic", "/health"),
};

export const studentApi = {
  create: (body) => request("students", "/students", { method: "POST", body }),
  list: (query) => request("students", "/students", { query }),
  update: (id, body) => request("students", `/students/${id}`, { method: "PATCH", body }),
  health: () => request("students", "/health"),
};

export const attendanceApi = {
  create: (body) => request("attendance", "/attendance", { method: "POST", body }),
  bulkCreate: (body) => request("attendance", "/attendance/bulk", { method: "POST", body }),
  list: (query) => request("attendance", "/attendance", { query }),
  update: (id, body) => request("attendance", `/attendance/${id}`, { method: "PATCH", body }),
  createTeacher: (body) =>
    request("attendance", "/teacher-attendance", { method: "POST", body }),
  bulkCreateTeacher: (body) =>
    request("attendance", "/teacher-attendance/bulk", { method: "POST", body }),
  listTeacher: (query) => request("attendance", "/teacher-attendance", { query }),
  updateTeacher: (id, body) =>
    request("attendance", `/teacher-attendance/${id}`, { method: "PATCH", body }),
  health: () => request("attendance", "/health"),
};

export const examApi = {
  createExam: (body) => request("exams", "/exams", { method: "POST", body }),
  enterMarks: (body) => request("exams", "/marks", { method: "POST", body }),
  publish: (body) => request("exams", "/results/publish", { method: "POST", body }),
  getResults: (query) => request("exams", "/results", { query }),
  health: () => request("exams", "/health"),
};

export const financeApi = {
  createFee: (body) => request("finance", "/fees", { method: "POST", body }),
  recordPayment: (body) => request("finance", "/payments", { method: "POST", body }),
  getDues: (query) => request("finance", "/dues", { query }),
  health: () => request("finance", "/health"),
};
