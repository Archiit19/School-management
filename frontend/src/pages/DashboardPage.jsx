import { useEffect, useState } from "react";
import { useAuth } from "../context/AuthContext";
import { authApi, rolesApi, academicApi, studentApi, attendanceApi, examApi, financeApi, transportApi } from "../api/client";

const SERVICES = [
  { name: "Auth", fn: authApi.health, port: 8081 },
  { name: "User / Roles", fn: rolesApi.health, port: 8082 },
  { name: "Academic", fn: academicApi.health, port: 8083 },
  { name: "Student", fn: studentApi.health, port: 8084 },
  { name: "Attendance", fn: attendanceApi.health, port: 8085 },
  { name: "Exam", fn: examApi.health, port: 8086 },
  { name: "Finance", fn: financeApi.health, port: 8087 },
  { name: "Transport", fn: transportApi.health, port: 8088 },
];

export default function DashboardPage() {
  const { user } = useAuth();
  const [health, setHealth] = useState({});

  useEffect(() => {
    SERVICES.forEach((s) => {
      s.fn()
        .then(() => setHealth((h) => ({ ...h, [s.name]: "up" })))
        .catch(() => setHealth((h) => ({ ...h, [s.name]: "down" })));
    });
  }, []);

  return (
    <>
      <div className="page-header">
        <h1>Dashboard</h1>
        <p>Welcome back, {user?.name || "Admin"}. Overview of all services.</p>
      </div>

      <div className="card">
        <div className="card-title">Your Profile</div>
        <div className="grid-3">
          <div className="form-group">
            <label>Name</label>
            <input readOnly value={user?.name || ""} />
          </div>
          <div className="form-group">
            <label>Email</label>
            <input readOnly value={user?.email || ""} />
          </div>
          <div className="form-group">
            <label>Role</label>
            <input readOnly value={user?.role_name || ""} />
          </div>
          <div className="form-group">
            <label>User ID</label>
            <input readOnly value={user?.id || ""} className="mono" />
          </div>
          <div className="form-group">
            <label>School ID</label>
            <input readOnly value={user?.school_id || ""} className="mono" />
          </div>
          <div className="form-group">
            <label>Active</label>
            <input readOnly value={user?.is_active ? "Yes" : "No"} />
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card-title">Service Health</div>
        <div className="stats-row">
          {SERVICES.map((s) => (
            <div className="stat-card" key={s.name}>
              <div className="label">{s.name} :{s.port}</div>
              <div className="value" style={{ color: health[s.name] === "up" ? "var(--clr-success)" : health[s.name] === "down" ? "var(--clr-danger)" : "var(--clr-text-secondary)" }}>
                {health[s.name] === "up" ? "Online" : health[s.name] === "down" ? "Offline" : "..."}
              </div>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}
