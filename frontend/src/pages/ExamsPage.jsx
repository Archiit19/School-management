import { useState, useEffect, useCallback } from "react";
import { examApi, academicApi } from "../api/client";

export default function ExamsPage() {
  const [tab, setTab] = useState("results");
  const [results, setResults] = useState([]);
  const [classes, setClasses] = useState([]);
  const [query, setQuery] = useState({ exam_id: "", student_id: "", class_id: "" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const [examForm, setExamForm] = useState({ class_id: "", section_id: "", subject_id: "", title: "", exam_date: "", total_marks: "" });
  const [marksForm, setMarksForm] = useState({ exam_id: "", student_id: "", marks_obtained: "", remarks: "" });
  const [publishForm, setPublishForm] = useState({ exam_id: "" });

  const load = useCallback(async () => {
    try { setResults(await examApi.getResults(query)); }
    catch (err) { setError(err.message); }
  }, [query]);

  useEffect(() => { load(); }, [load]);
  useEffect(() => { academicApi.getClasses().then(setClasses).catch(() => {}); }, []);

  const flatClasses = classes.map((c) => c.class || c);
  const flatSubjects = classes.flatMap((c) => (c.subjects || []).map((s) => ({ ...s, className: (c.class || c).name })));
  const flatSections = classes.flatMap((c) => (c.sections || []).map((s) => ({ ...s, className: (c.class || c).name })));

  function msg(txt) { setSuccess(txt); setError(""); setTimeout(() => setSuccess(""), 3000); }

  async function handleCreateExam(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try {
      const payload = { ...examForm, total_marks: parseFloat(examForm.total_marks) };
      if (!payload.section_id) delete payload.section_id;
      if (!payload.subject_id) delete payload.subject_id;
      await examApi.createExam(payload);
      msg("Exam created.");
      setExamForm({ class_id: "", section_id: "", subject_id: "", title: "", exam_date: "", total_marks: "" });
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  async function handleEnterMarks(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try {
      await examApi.enterMarks({ ...marksForm, marks_obtained: parseFloat(marksForm.marks_obtained) });
      msg("Marks saved.");
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  async function handlePublish(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try { await examApi.publish(publishForm); msg("Results published."); load(); }
    catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  return (
    <>
      <div className="page-header">
        <h1>Exams & Results</h1>
        <p>Flow 8 — Create exams, enter marks, publish, and view results.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="tabs">
        <button className={`tab ${tab === "results" ? "active" : ""}`} onClick={() => setTab("results")}>Results</button>
        <button className={`tab ${tab === "exam" ? "active" : ""}`} onClick={() => setTab("exam")}>Create Exam</button>
        <button className={`tab ${tab === "marks" ? "active" : ""}`} onClick={() => setTab("marks")}>Enter Marks</button>
        <button className={`tab ${tab === "publish" ? "active" : ""}`} onClick={() => setTab("publish")}>Publish</button>
      </div>

      {tab === "results" && (
        <div className="card">
          <div className="card-title">Results <span className="badge badge-get">GET /results</span></div>
          <div className="grid-3 mb-4">
            <div className="form-group"><label>Exam ID</label><input placeholder="UUID..." value={query.exam_id} onChange={(e) => setQuery((q) => ({ ...q, exam_id: e.target.value }))} /></div>
            <div className="form-group"><label>Student ID</label><input placeholder="UUID..." value={query.student_id} onChange={(e) => setQuery((q) => ({ ...q, student_id: e.target.value }))} /></div>
            <div className="form-group">
              <label>Class</label>
              <select value={query.class_id} onChange={(e) => setQuery((q) => ({ ...q, class_id: e.target.value }))}>
                <option value="">All</option>
                {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </div>
          </div>
          <div className="table-wrap">
            <table>
              <thead><tr><th>Exam</th><th>Date</th><th>Student ID</th><th>Marks</th><th>Total</th><th>%</th><th>Grade</th></tr></thead>
              <tbody>
                {(!results || results.length === 0) && <tr><td colSpan={7} className="empty">No results.</td></tr>}
                {(results || []).map((r, i) => (
                  <tr key={i}>
                    <td><strong>{r.exam_title}</strong></td>
                    <td>{r.exam_date?.split("T")[0]}</td>
                    <td><span className="mono truncate">{r.student_id}</span></td>
                    <td>{r.marks_obtained}</td>
                    <td>{r.total_marks}</td>
                    <td>{r.percentage?.toFixed(1)}%</td>
                    <td><span className="status status-active">{r.grade}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {tab === "exam" && (
        <div className="card">
          <div className="card-title">Create Exam <span className="badge badge-post">POST /exams</span></div>
          <form onSubmit={handleCreateExam}>
            <div className="grid-3">
              <div className="form-group"><label>Title</label><input required value={examForm.title} onChange={(e) => setExamForm((p) => ({ ...p, title: e.target.value }))} placeholder="Mid-term Math" /></div>
              <div className="form-group">
                <label>Class</label>
                <select required value={examForm.class_id} onChange={(e) => setExamForm((p) => ({ ...p, class_id: e.target.value }))}>
                  <option value="">Select...</option>
                  {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Subject (optional)</label>
                <select value={examForm.subject_id} onChange={(e) => setExamForm((p) => ({ ...p, subject_id: e.target.value }))}>
                  <option value="">Any</option>
                  {flatSubjects.map((s) => <option key={s.id} value={s.id}>{s.className} — {s.name}</option>)}
                </select>
              </div>
            </div>
            <div className="grid-3">
              <div className="form-group">
                <label>Section (optional)</label>
                <select value={examForm.section_id} onChange={(e) => setExamForm((p) => ({ ...p, section_id: e.target.value }))}>
                  <option value="">Any</option>
                  {flatSections.map((s) => <option key={s.id} value={s.id}>{s.className} — {s.name}</option>)}
                </select>
              </div>
              <div className="form-group"><label>Exam Date</label><input type="date" required value={examForm.exam_date} onChange={(e) => setExamForm((p) => ({ ...p, exam_date: e.target.value }))} /></div>
              <div className="form-group"><label>Total Marks</label><input type="number" required min="1" value={examForm.total_marks} onChange={(e) => setExamForm((p) => ({ ...p, total_marks: e.target.value }))} placeholder="100" /></div>
            </div>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Create Exam</button></div>
          </form>
        </div>
      )}

      {tab === "marks" && (
        <div className="card">
          <div className="card-title">Enter Marks <span className="badge badge-post">POST /marks</span></div>
          <form onSubmit={handleEnterMarks}>
            <div className="grid-4">
              <div className="form-group"><label>Exam ID</label><input required value={marksForm.exam_id} onChange={(e) => setMarksForm((p) => ({ ...p, exam_id: e.target.value }))} placeholder="UUID" /></div>
              <div className="form-group"><label>Student ID</label><input required value={marksForm.student_id} onChange={(e) => setMarksForm((p) => ({ ...p, student_id: e.target.value }))} placeholder="UUID" /></div>
              <div className="form-group"><label>Marks Obtained</label><input type="number" required min="0" value={marksForm.marks_obtained} onChange={(e) => setMarksForm((p) => ({ ...p, marks_obtained: e.target.value }))} placeholder="85" /></div>
              <div className="form-group"><label>Remarks</label><input value={marksForm.remarks} onChange={(e) => setMarksForm((p) => ({ ...p, remarks: e.target.value }))} placeholder="Optional" /></div>
            </div>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Save Marks</button></div>
          </form>
        </div>
      )}

      {tab === "publish" && (
        <div className="card">
          <div className="card-title">Publish Results <span className="badge badge-post">POST /results/publish</span></div>
          <form onSubmit={handlePublish}>
            <div className="form-group" style={{ maxWidth: 400 }}>
              <label>Exam ID</label>
              <input required value={publishForm.exam_id} onChange={(e) => setPublishForm({ exam_id: e.target.value })} placeholder="UUID of exam to publish" />
            </div>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Publish</button></div>
          </form>
        </div>
      )}
    </>
  );
}
