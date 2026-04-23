import { useState, useEffect, useCallback } from "react";
import { academicApi } from "../api/client";

export default function TeacherAssignmentsPage() {
  const [assignments, setAssignments] = useState([]);
  const [classes, setClasses] = useState([]);
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

  const flatClasses = classes.map((c) => c.class || c);
  const flatSubjects = classes.flatMap((c) => (c.subjects || []).map((s) => ({ ...s, className: (c.class || c).name })));

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
        <p>Flow 5 — Assign teachers to class + subject combinations.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="card">
        <div className="card-title">Assign Teacher <span className="badge badge-post">POST /teacher-assignments</span></div>
        <form onSubmit={handleCreate}>
          <div className="grid-3">
            <div className="form-group"><label>Teacher User ID</label><input name="teacher_user_id" required value={form.teacher_user_id} onChange={field} placeholder="UUID of teacher user" /></div>
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
        <div className="card-title">Assignments <span className="badge badge-get">GET /teacher-assignments</span></div>
        <div className="grid-3 mb-4">
          <div className="form-group"><label>Teacher ID</label><input placeholder="Filter by teacher UUID..." value={query.teacher_user_id} onChange={(e) => setQuery((q) => ({ ...q, teacher_user_id: e.target.value }))} /></div>
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
            <thead><tr><th>Teacher User ID</th><th>Class ID</th><th>Subject ID</th><th>ID</th></tr></thead>
            <tbody>
              {assignments.length === 0 && <tr><td colSpan={4} className="empty">No assignments found.</td></tr>}
              {assignments.map((a) => (
                <tr key={a.id}>
                  <td><span className="mono truncate">{a.teacher_user_id}</span></td>
                  <td><span className="mono truncate">{a.class_id}</span></td>
                  <td><span className="mono truncate">{a.subject_id}</span></td>
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
