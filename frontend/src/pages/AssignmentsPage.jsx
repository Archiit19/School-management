import { useState, useEffect, useCallback } from "react";
import { academicApi } from "../api/client";

export default function AssignmentsPage() {
  const [tab, setTab] = useState("list");
  const [assignments, setAssignments] = useState([]);
  const [classes, setClasses] = useState([]);
  const [query, setQuery] = useState({ class_id: "", subject_id: "", teacher_id: "" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const [form, setForm] = useState({ class_id: "", subject_id: "", title: "", description: "", material_url: "", due_date: "" });
  const [subForm, setSubForm] = useState({ assignment_id: "", student_id: "", content: "", material_url: "" });

  const load = useCallback(async () => {
    try { setAssignments(await academicApi.getAssignments(query)); }
    catch (err) { setError(err.message); }
  }, [query]);

  useEffect(() => { load(); }, [load]);
  useEffect(() => { academicApi.getClasses().then(setClasses).catch(() => {}); }, []);

  const flatClasses = classes.map((c) => c.class || c);
  const flatSubjects = classes.flatMap((c) => (c.subjects || []).map((s) => ({ ...s, className: (c.class || c).name })));

  function msg(txt) { setSuccess(txt); setError(""); setTimeout(() => setSuccess(""), 3000); }

  async function handleCreate(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try {
      const payload = { ...form };
      if (!payload.due_date) delete payload.due_date;
      await academicApi.createAssignment(payload);
      msg("Assignment created.");
      setForm({ class_id: "", subject_id: "", title: "", description: "", material_url: "", due_date: "" });
      load();
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  async function handleSubmit(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try {
      await academicApi.createSubmission(subForm);
      msg("Submission recorded.");
      setSubForm({ assignment_id: "", student_id: "", content: "", material_url: "" });
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  return (
    <>
      <div className="page-header">
        <h1>Assignments & Submissions</h1>
        <p>Flow 7 — Create assignments and record student submissions.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="tabs">
        <button className={`tab ${tab === "list" ? "active" : ""}`} onClick={() => setTab("list")}>View</button>
        <button className={`tab ${tab === "create" ? "active" : ""}`} onClick={() => setTab("create")}>Create Assignment</button>
        <button className={`tab ${tab === "submit" ? "active" : ""}`} onClick={() => setTab("submit")}>Submit Work</button>
      </div>

      {tab === "list" && (
        <div className="card">
          <div className="card-title">Assignments <span className="badge badge-get">GET /assignments</span></div>
          <div className="grid-3 mb-4">
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
            <div className="form-group"><label>Teacher ID</label><input placeholder="UUID..." value={query.teacher_id} onChange={(e) => setQuery((q) => ({ ...q, teacher_id: e.target.value }))} /></div>
          </div>
          <div className="table-wrap">
            <table>
              <thead><tr><th>Title</th><th>Description</th><th>Due Date</th><th>Teacher ID</th><th>ID</th></tr></thead>
              <tbody>
                {assignments.length === 0 && <tr><td colSpan={5} className="empty">No assignments.</td></tr>}
                {assignments.map((a) => (
                  <tr key={a.id}>
                    <td><strong>{a.title}</strong></td>
                    <td>{a.description || "—"}</td>
                    <td>{a.due_date ? a.due_date.split("T")[0] : "—"}</td>
                    <td><span className="mono truncate">{a.teacher_user_id}</span></td>
                    <td><span className="mono truncate">{a.id}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {tab === "create" && (
        <div className="card">
          <div className="card-title">Create Assignment <span className="badge badge-post">POST /assignments</span></div>
          <form onSubmit={handleCreate}>
            <div className="grid-3">
              <div className="form-group">
                <label>Class</label>
                <select required value={form.class_id} onChange={(e) => setForm((p) => ({ ...p, class_id: e.target.value }))}>
                  <option value="">Select...</option>
                  {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Subject</label>
                <select required value={form.subject_id} onChange={(e) => setForm((p) => ({ ...p, subject_id: e.target.value }))}>
                  <option value="">Select...</option>
                  {flatSubjects.map((s) => <option key={s.id} value={s.id}>{s.className} — {s.name}</option>)}
                </select>
              </div>
              <div className="form-group"><label>Title</label><input required value={form.title} onChange={(e) => setForm((p) => ({ ...p, title: e.target.value }))} placeholder="Homework #1" /></div>
            </div>
            <div className="grid-3">
              <div className="form-group"><label>Description</label><input value={form.description} onChange={(e) => setForm((p) => ({ ...p, description: e.target.value }))} placeholder="Details" /></div>
              <div className="form-group"><label>Material URL</label><input value={form.material_url} onChange={(e) => setForm((p) => ({ ...p, material_url: e.target.value }))} placeholder="https://..." /></div>
              <div className="form-group"><label>Due Date</label><input type="date" value={form.due_date} onChange={(e) => setForm((p) => ({ ...p, due_date: e.target.value }))} /></div>
            </div>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Create</button></div>
          </form>
        </div>
      )}

      {tab === "submit" && (
        <div className="card">
          <div className="card-title">Submit Student Work <span className="badge badge-post">POST /submissions</span></div>
          <form onSubmit={handleSubmit}>
            <div className="grid-2">
              <div className="form-group"><label>Assignment ID</label><input required value={subForm.assignment_id} onChange={(e) => setSubForm((p) => ({ ...p, assignment_id: e.target.value }))} placeholder="UUID" /></div>
              <div className="form-group"><label>Student ID</label><input required value={subForm.student_id} onChange={(e) => setSubForm((p) => ({ ...p, student_id: e.target.value }))} placeholder="UUID" /></div>
            </div>
            <div className="grid-2">
              <div className="form-group"><label>Content</label><textarea value={subForm.content} onChange={(e) => setSubForm((p) => ({ ...p, content: e.target.value }))} placeholder="Student answer..." /></div>
              <div className="form-group"><label>Material URL</label><input value={subForm.material_url} onChange={(e) => setSubForm((p) => ({ ...p, material_url: e.target.value }))} placeholder="https://..." /></div>
            </div>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Submit</button></div>
          </form>
        </div>
      )}
    </>
  );
}
