import { useState, useEffect, useCallback, useMemo } from "react";
import { academicApi, rolesApi, userMgmtApi } from "../api/client";
import PermGate from "../components/PermGate";
import { useAuth } from "../context/AuthContext";

export default function TeacherAssignmentsPage() {
  const { hasPerm } = useAuth();
  const [assignments, setAssignments] = useState([]);
  const [classes, setClasses] = useState([]);
  const [teachers, setTeachers] = useState([]);
  const [teachersLoading, setTeachersLoading] = useState(false);
  const [teachersError, setTeachersError] = useState("");
  const [query, setQuery] = useState({ teacher_user_id: "", class_id: "", subject_id: "" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const [form, setForm] = useState({ teacher_user_id: "", class_id: "", subject_id: "" });

  const load = useCallback(async () => {
    try { setAssignments(await academicApi.getTeacherAssignments(query)); }
    catch (err) { setError(err.message); }
  }, [query]);

  useEffect(() => { load(); }, [load]);
  useEffect(() => { academicApi.getClasses().then(setClasses).catch(() => {}); }, []);

  useEffect(() => {
    if (!hasPerm("view_users")) {
      setTeachersLoading(false);
      return;
    }
    let cancelled = false;
    (async () => {
      setTeachersLoading(true);
      setTeachersError("");
      try {
        const roles = await rolesApi.list();
        const teacherRole = roles.find((r) => r.name === "teacher");
        if (!teacherRole) {
          if (!cancelled) setTeachers([]);
          return;
        }
        const res = await userMgmtApi.list({ role_id: teacherRole.id, limit: 200, page: 1 });
        if (!cancelled) setTeachers(res.users || []);
      } catch (e) {
        if (!cancelled) setTeachersError(e.message);
      } finally {
        if (!cancelled) setTeachersLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, [hasPerm]);

  const flatClasses = classes.map((c) => c.class || c);
  const flatSubjects = classes.flatMap((c) => (c.subjects || []).map((s) => ({ ...s, className: (c.class || c).name })));

  const classMap = Object.fromEntries(flatClasses.map((c) => [c.id, c.name]));
  const subjectMap = Object.fromEntries(flatSubjects.map((s) => [s.id, s.name]));
  const teacherMap = useMemo(
    () => Object.fromEntries(teachers.map((t) => [t.id, t])),
    [teachers],
  );

  function teacherLabel(id) {
    const t = teacherMap[id];
    if (t) return `${t.name} (${t.email})`;
    return id;
  }

  function field(e) { setForm((p) => ({ ...p, [e.target.name]: e.target.value })); }
  function msg(txt) { setSuccess(txt); setError(""); setTimeout(() => setSuccess(""), 3000); }

  async function handleCreate(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try { await academicApi.createTeacherAssignment(form); msg("Teacher assigned."); load(); }
    catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  return (
    <>
      <div className="page-header">
        <h1>Teacher Assignments</h1>
        <p>Assign teachers to class + subject. Pick a teacher from the list below (create teachers under User Management first).</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <PermGate perm="view_users">
        <div className="card">
          <div className="card-title">
            School teachers <span className="badge badge-get">GET /users?role_id=…</span>
          </div>
          {teachersError && <div className="alert alert-error">{teachersError}</div>}
          <div className="table-wrap">
            <table>
              <thead><tr><th>Name</th><th>Email</th><th>User ID</th><th>Status</th></tr></thead>
              <tbody>
                {teachersLoading && <tr><td colSpan={4} className="empty">Loading teachers…</td></tr>}
                {!teachersLoading && teachers.length === 0 && (
                  <tr><td colSpan={4} className="empty">No teacher accounts yet. Create users with the &quot;teacher&quot; role under User Management.</td></tr>
                )}
                {!teachersLoading && teachers.map((t) => (
                  <tr key={t.id}>
                    <td><strong>{t.name}</strong></td>
                    <td>{t.email}</td>
                    <td><span className="mono truncate">{t.id}</span></td>
                    <td>
                      <span className={`status ${t.is_active ? "status-active" : "status-inactive"}`}>
                        {t.is_active ? "Active" : "Inactive"}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </PermGate>

      <div className="card">
        <div className="card-title">Assign Teacher <span className="badge badge-post">POST /teacher-assignments</span></div>
        <form onSubmit={handleCreate}>
          <div className="grid-3">
            <div className="form-group">
              <label>Teacher</label>
              {hasPerm("view_users") && teachers.length > 0 ? (
                <select name="teacher_user_id" required value={form.teacher_user_id} onChange={field}>
                  <option value="">Select teacher…</option>
                  {teachers.filter((t) => t.is_active !== false).map((t) => (
                    <option key={t.id} value={t.id}>{t.name} — {t.email}</option>
                  ))}
                </select>
              ) : (
                <input name="teacher_user_id" required value={form.teacher_user_id} onChange={field} placeholder="UUID of teacher user" />
              )}
            </div>
            <div className="form-group">
              <label>Class</label>
              <select name="class_id" required value={form.class_id} onChange={field}>
                <option value="">Select class...</option>
                {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label>Subject</label>
              <select name="subject_id" required value={form.subject_id} onChange={field}>
                <option value="">Select subject...</option>
                {flatSubjects.map((s) => <option key={s.id} value={s.id}>{s.className} — {s.name}</option>)}
              </select>
            </div>
          </div>
          <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Assign</button></div>
        </form>
      </div>

      <div className="card">
        <div className="card-title">Class assignments <span className="badge badge-get">GET /teacher-assignments</span></div>
        <div className="grid-3 mb-4">
          <div className="form-group">
            <label>Teacher</label>
            {hasPerm("view_users") && teachers.length > 0 ? (
              <select value={query.teacher_user_id} onChange={(e) => setQuery((q) => ({ ...q, teacher_user_id: e.target.value }))}>
                <option value="">All teachers</option>
                {teachers.map((t) => (
                  <option key={t.id} value={t.id}>{t.name}</option>
                ))}
              </select>
            ) : (
              <input placeholder="Filter by teacher UUID..." value={query.teacher_user_id} onChange={(e) => setQuery((q) => ({ ...q, teacher_user_id: e.target.value }))} />
            )}
          </div>
          <div className="form-group">
            <label>Class</label>
            <select value={query.class_id} onChange={(e) => setQuery((q) => ({ ...q, class_id: e.target.value }))}>
              <option value="">All</option>
              {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </div>
          <div className="form-group">
            <label>Subject</label>
            <select value={query.subject_id} onChange={(e) => setQuery((q) => ({ ...q, subject_id: e.target.value }))}>
              <option value="">All</option>
              {flatSubjects.map((s) => <option key={s.id} value={s.id}>{s.className} — {s.name}</option>)}
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Teacher</th><th>Class</th><th>Subject</th><th>Assignment ID</th></tr></thead>
            <tbody>
              {assignments.length === 0 && <tr><td colSpan={4} className="empty">No assignments found.</td></tr>}
              {assignments.map((a) => (
                <tr key={a.id}>
                  <td>
                    <strong>{teacherMap[a.teacher_user_id]?.name || "—"}</strong>
                    {teacherMap[a.teacher_user_id]?.email && (
                      <div className="text-sm text-muted">{teacherMap[a.teacher_user_id].email}</div>
                    )}
                    {!teacherMap[a.teacher_user_id] && (
                      <span className="mono truncate text-sm">{a.teacher_user_id}</span>
                    )}
                  </td>
                  <td>{classMap[a.class_id] || a.class_id}</td>
                  <td>{subjectMap[a.subject_id] || a.subject_id}</td>
                  <td><span className="mono truncate">{a.id}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
