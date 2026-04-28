const SERVICES = {
  auth: "/api/auth",
  users: "/api/users",
  academic: "/api/academic",
  students: "/api/students",
  attendance: "/api/attendance",
  exams: "/api/exams",
  finance: "/api/finance",
  transport: "/api/transport",
};

function getToken() {
  return localStorage.getItem("token") || "";
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
  stats: (query) => request("attendance", "/attendance/stats", { query }),
  createTeacher: (body) =>
    request("attendance", "/teacher-attendance", { method: "POST", body }),
  bulkCreateTeacher: (body) =>
    request("attendance", "/teacher-attendance/bulk", { method: "POST", body }),
  listTeacher: (query) => request("attendance", "/teacher-attendance", { query }),
  updateTeacher: (id, body) =>
    request("attendance", `/teacher-attendance/${id}`, { method: "PATCH", body }),
  statsTeacher: (query) => request("attendance", "/teacher-attendance/stats", { query }),
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

export const transportApi = {
  // Vehicles
  createVehicle: (body) => request("transport", "/vehicles", { method: "POST", body }),
  getVehicles: (query) => request("transport", "/vehicles", { query }),
  getVehicle: (id) => request("transport", `/vehicles/${id}`),
  updateVehicle: (id, body) => request("transport", `/vehicles/${id}`, { method: "PATCH", body }),
  deleteVehicle: (id) => request("transport", `/vehicles/${id}`, { method: "DELETE" }),
  // Routes
  createRoute: (body) => request("transport", "/routes", { method: "POST", body }),
  getRoutes: (query) => request("transport", "/routes", { query }),
  getRoute: (id) => request("transport", `/routes/${id}`),
  updateRoute: (id, body) => request("transport", `/routes/${id}`, { method: "PATCH", body }),
  deleteRoute: (id) => request("transport", `/routes/${id}`, { method: "DELETE" }),
  // Stops
  createStop: (body) => request("transport", "/stops", { method: "POST", body }),
  getStops: (query) => request("transport", "/stops", { query }),
  getStop: (id) => request("transport", `/stops/${id}`),
  updateStop: (id, body) => request("transport", `/stops/${id}`, { method: "PATCH", body }),
  deleteStop: (id) => request("transport", `/stops/${id}`, { method: "DELETE" }),
  // Student Transport
  assignStudent: (body) => request("transport", "/student-transport", { method: "POST", body }),
  getStudentTransports: (query) => request("transport", "/student-transport", { query }),
  getStudentTransport: (id) => request("transport", `/student-transport/${id}`),
  updateStudentTransport: (id, body) => request("transport", `/student-transport/${id}`, { method: "PATCH", body }),
  deleteStudentTransport: (id) => request("transport", `/student-transport/${id}`, { method: "DELETE" }),
  health: () => request("transport", "/health"),
};
