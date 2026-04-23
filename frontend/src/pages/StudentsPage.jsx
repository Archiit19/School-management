import { useState, useEffect, useCallback } from "react";
import { studentApi, academicApi } from "../api/client";

export default function StudentsPage() {
  const [students, setStudents] = useState([]);
  const [classes, setClasses] = useState([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({ page: 1, limit: 20, search: "", class_id: "", section_id: "" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const [form, setForm] = useState({ first_name: "", last_name: "", class_id: "", section_id: "", parent_user_id: "" });

  const load = useCallback(async () => {
    try {
      const res = await studentApi.list(query);
      setStudents(res.students || []);
      setTotal(res.total || 0);
    } catch (err) { setError(err.message); }
  }, [query]);

  useEffect(() => { load(); }, [load]);
  useEffect(() => { academicApi.getClasses().then(setClasses).catch(() => {}); }, []);

  const flatClasses = classes.map((c) => c.class || c);
  const flatSections = classes.flatMap((c) => (c.sections || []).map((s) => ({ ...s, className: (c.class || c).name })));

  function field(e) { setForm((p) => ({ ...p, [e.target.name]: e.target.value })); }

  function msg(txt) { setSuccess(txt); setError(""); setTimeout(() => setSuccess(""), 3000); }

  async function handleCreate(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try {
      const payload = { ...form };
      if (!payload.section_id) delete payload.section_id;
      if (!payload.parent_user_id) delete payload.parent_user_id;
      await studentApi.create(payload);
      msg("Student admitted.");
      setForm({ first_name: "", last_name: "", class_id: "", section_id: "", parent_user_id: "" });
      load();
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  return (
    <>
      <div className="page-header">
        <h1>Students</h1>
        <p>Flow 4 — Admit and manage students.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="card">
        <div className="card-title">Admit Student <span className="badge badge-post">POST /students</span></div>
        <form onSubmit={handleCreate}>
          <div className="grid-3">
            <div className="form-group"><label>First Name</label><input name="first_name" required value={form.first_name} onChange={field} placeholder="John" /></div>
            <div className="form-group"><label>Last Name</label><input name="last_name" required value={form.last_name} onChange={field} placeholder="Doe" /></div>
            <div className="form-group">
              <label>Class</label>
              <select name="class_id" required value={form.class_id} onChange={field}>
                <option value="">Select class...</option>
                {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label>Section (optional)</label>
              <select name="section_id" value={form.section_id} onChange={field}>
                <option value="">Any</option>
                {flatSections.map((s) => <option key={s.id} value={s.id}>{s.className} — {s.name}</option>)}
              </select>
            </div>
            <div className="form-group"><label>Parent User ID (optional)</label><input name="parent_user_id" value={form.parent_user_id} onChange={field} placeholder="UUID of parent user" /></div>
          </div>
          <div className="btn-row"><button className="btn btn-primary" disabled={busy}>{busy ? "Admitting..." : "Admit Student"}</button></div>
        </form>
      </div>

      <div className="card">
        <div className="card-title">Students ({total}) <span className="badge badge-get">GET /students</span></div>
        <div className="grid-3 mb-4">
          <div className="form-group"><label>Search</label><input placeholder="Name..." value={query.search} onChange={(e) => setQuery((q) => ({ ...q, search: e.target.value, page: 1 }))} /></div>
          <div className="form-group">
            <label>Class</label>
            <select value={query.class_id} onChange={(e) => setQuery((q) => ({ ...q, class_id: e.target.value, page: 1 }))}>
              <option value="">All</option>
              {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </div>
          <div className="form-group">
            <label>Section</label>
            <select value={query.section_id} onChange={(e) => setQuery((q) => ({ ...q, section_id: e.target.value, page: 1 }))}>
              <option value="">All</option>
              {flatSections.map((s) => <option key={s.id} value={s.id}>{s.className} — {s.name}</option>)}
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Name</th><th>Class ID</th><th>Section ID</th><th>Active</th><th>Student ID</th></tr></thead>
            <tbody>
              {students.length === 0 && <tr><td colSpan={5} className="empty">No students found.</td></tr>}
              {students.map((s) => (
                <tr key={s.id}>
                  <td>{s.first_name} {s.last_name}</td>
                  <td><span className="mono truncate">{s.class_id}</span></td>
                  <td><span className="mono truncate">{s.section_id || "—"}</span></td>
                  <td><span className={`status ${s.is_active ? "status-active" : "status-inactive"}`}>{s.is_active ? "Active" : "Inactive"}</span></td>
                  <td><span className="mono truncate">{s.id}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {total > query.limit && (
          <div className="btn-row">
            <button className="btn btn-ghost btn-sm" disabled={query.page <= 1} onClick={() => setQuery((q) => ({ ...q, page: q.page - 1 }))}>Prev</button>
            <span className="text-sm text-muted" style={{ padding: "6px" }}>Page {query.page}</span>
            <button className="btn btn-ghost btn-sm" disabled={query.page * query.limit >= total} onClick={() => setQuery((q) => ({ ...q, page: q.page + 1 }))}>Next</button>
          </div>
        )}
      </div>
    </>
  );
}
