import { useState, useEffect, useCallback, useMemo } from "react";
import { examApi, academicApi } from "../api/client";
import PermTabBar from "../components/PermTabBar";
import { usePermTabs } from "../hooks/usePermTabs";

function classNodeId(node) {
  return (node.class || node).id;
}

const EXAM_TABS = [
  { id: "exams", label: "Exams", perm: "view_exams" },
  { id: "results", label: "Results", perm: "view_results" },
  { id: "exam", label: "Create Exam", perm: "create_exam" },
  { id: "marks", label: "Enter Marks", perm: "enter_marks" },
  { id: "publish", label: "Publish", perm: "publish_results" },
];

export default function ExamsPage() {
  const { visibleTabs, tab, setTab } = usePermTabs(EXAM_TABS, "exams");
  const [results, setResults] = useState([]);
  const [exams, setExams] = useState([]);
  const [classes, setClasses] = useState([]);
  const [query, setQuery] = useState({ exam_id: "", student_id: "", class_id: "" });
  const [examFilter, setExamFilter] = useState({ class_id: "", section_id: "", upcoming: true });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const [examForm, setExamForm] = useState({ class_id: "", section_id: "", subject_id: "", title: "", exam_date: "", total_marks: "" });
  const [marksForm, setMarksForm] = useState({ exam_id: "", student_id: "", marks_obtained: "", remarks: "" });
  const [publishForm, setPublishForm] = useState({ exam_id: "" });

  const loadExams = useCallback(async () => {
    try {
      const q = { ...examFilter };
      if (!q.class_id) delete q.class_id;
      if (!q.section_id) delete q.section_id;
      setExams(await examApi.getExams(q));
    } catch (err) { setError(err.message); }
  }, [examFilter]);

  const loadResults = useCallback(async () => {
    try { setResults(await examApi.getResults(query)); }
    catch (err) { setError(err.message); }
  }, [query]);

  const [teacherAssignments, setTeacherAssignments] = useState([]);

  useEffect(() => { if (tab === "exams" || tab === "exam" || tab === "publish") loadExams(); }, [loadExams, tab]);
  useEffect(() => { if (tab === "results") loadResults(); }, [loadResults, tab]);
  useEffect(() => { academicApi.getClasses().then(setClasses).catch(() => {}); }, []);

  useEffect(() => {
    if (!examForm.class_id) {
      setTeacherAssignments([]);
      return;
    }
    academicApi
      .getTeacherAssignments({ class_id: examForm.class_id })
      .then(setTeacherAssignments)
      .catch(() => setTeacherAssignments([]));
  }, [examForm.class_id]);

  const flatClasses = classes.map((c) => c.class || c);
  const flatSubjects = classes.flatMap((c) => (c.subjects || []).map((s) => ({ ...s, class_id: (c.class || c).id, className: (c.class || c).name })));
  const flatSections = classes.flatMap((c) => (c.sections || []).map((s) => ({ ...s, class_id: (c.class || c).id, className: (c.class || c).name })));

  const examFormClassNode = useMemo(
    () => classes.find((c) => classNodeId(c) === examForm.class_id),
    [classes, examForm.class_id],
  );
  const examFormSections = examFormClassNode?.sections || [];
  const examFormHasSections = examFormSections.length > 0;

  const assignedSubjectIdsForClass = useMemo(
    () => new Set(teacherAssignments.map((ta) => ta.subject_id)),
    [teacherAssignments],
  );

  const examFormAvailableSubjects = useMemo(() => {
    if (!examForm.class_id || !examFormClassNode) return [];
    const subjects = examFormClassNode.subjects || [];
    return subjects.filter((s) => {
      if (!assignedSubjectIdsForClass.has(s.id)) return false;
      if (examFormHasSections) {
        if (!examForm.section_id) return false;
        if (!s.section_id) return true;
        return s.section_id === examForm.section_id;
      }
      return !s.section_id;
    });
  }, [examForm.class_id, examForm.section_id, examFormClassNode, examFormHasSections, assignedSubjectIdsForClass]);

  const canPickExamSubject = examForm.class_id && (!examFormHasSections || examForm.section_id);

  function onExamFormClassChange(classId) {
    setExamForm((p) => ({ ...p, class_id: classId, section_id: "", subject_id: "" }));
  }

  function onExamFormSectionChange(sectionId) {
    setExamForm((p) => ({ ...p, section_id: sectionId, subject_id: "" }));
  }

  function msg(txt) { setSuccess(txt); setError(""); setTimeout(() => setSuccess(""), 3000); }

  function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => msg("Copied to clipboard!")).catch(() => setError("Failed to copy"));
  }

  async function handleCreateExam(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try {
      const payload = { ...examForm, total_marks: parseFloat(examForm.total_marks) };
      if (!payload.section_id) delete payload.section_id;
      if (!payload.subject_id) delete payload.subject_id;
      const created = await examApi.createExam(payload);
      msg(`Exam created! ID: ${created.id}`);
      setExamForm({ class_id: "", section_id: "", subject_id: "", title: "", exam_date: "", total_marks: "" });
      loadExams();
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  const classNameById = new Map(classes.map((c) => { const cls = c.class || c; return [cls.id, cls.name]; }));
  const sectionNameById = new Map(classes.flatMap((c) => (c.sections || []).map((s) => [s.id, s.name])));
  const subjectNameById = new Map(classes.flatMap((c) => (c.subjects || []).map((s) => [s.id, s.name])));

  async function handleEnterMarks(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try {
      await examApi.enterMarks({ ...marksForm, marks_obtained: parseFloat(marksForm.marks_obtained) });
      msg("Marks saved.");
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  async function handlePublish(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try { await examApi.publish(publishForm); msg("Results published."); loadResults(); loadExams(); }
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

      <PermTabBar tabs={visibleTabs} active={tab} onChange={setTab} />

      {tab === "exams" && (
        <div className="card">
          <div className="card-title">Exams <span className="badge badge-get">GET /exams</span></div>
          <div className="grid-3 mb-4">
            <div className="form-group">
              <label>Class</label>
              <select value={examFilter.class_id} onChange={(e) => setExamFilter((q) => ({ ...q, class_id: e.target.value, section_id: "" }))}>
                <option value="">All classes</option>
                {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label>Section</label>
              <select value={examFilter.section_id} onChange={(e) => setExamFilter((q) => ({ ...q, section_id: e.target.value }))}>
                <option value="">Any section</option>
                {flatSections
                  .filter((s) => !examFilter.class_id || s.class_id === examFilter.class_id)
                  .map((s) => <option key={s.id} value={s.id}>{s.className} — {s.name}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label style={{ display: "flex", alignItems: "center", gap: 8 }}>
                <input type="checkbox" checked={examFilter.upcoming} onChange={(e) => setExamFilter((q) => ({ ...q, upcoming: e.target.checked }))} />
                Upcoming only
              </label>
            </div>
          </div>
          <div className="table-wrap">
            <table>
              <thead><tr><th>Title</th><th>Class</th><th>Section</th><th>Subject</th><th>Date</th><th>Total</th><th>Status</th><th>Exam ID</th></tr></thead>
              <tbody>
                {(!exams || exams.length === 0) && <tr><td colSpan={8} className="empty">No exams found.</td></tr>}
                {(exams || []).map((e) => (
                  <tr key={e.id}>
                    <td><strong>{e.title}</strong></td>
                    <td>{classNameById.get(e.class_id) || <span className="mono truncate">{e.class_id}</span>}</td>
                    <td>{e.section_id ? (sectionNameById.get(e.section_id) || <span className="mono truncate">{e.section_id}</span>) : "—"}</td>
                    <td>{e.subject_id ? (subjectNameById.get(e.subject_id) || <span className="mono truncate">{e.subject_id}</span>) : "—"}</td>
                    <td>{e.exam_date?.split("T")[0]}</td>
                    <td>{e.total_marks}</td>
                    <td><span className={`status ${e.is_published ? "status-active" : "status-inactive"}`}>{e.is_published ? "Published" : "Draft"}</span></td>
                    <td><span className="mono truncate">{e.id}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

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
              <div className="form-group">
                <label>Title</label>
                <input required value={examForm.title} onChange={(e) => setExamForm((p) => ({ ...p, title: e.target.value }))} placeholder="Mid-term Math" />
              </div>
              <div className="form-group">
                <label>Class</label>
                <select required value={examForm.class_id} onChange={(e) => onExamFormClassChange(e.target.value)}>
                  <option value="">Select class…</option>
                  {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Section</label>
                <select
                  required={examFormHasSections}
                  disabled={!examForm.class_id}
                  value={examForm.section_id}
                  onChange={(e) => onExamFormSectionChange(e.target.value)}
                >
                  <option value="">
                    {!examForm.class_id
                      ? "Select class first…"
                      : examFormHasSections
                        ? "Select section…"
                        : "No sections (class-wide)"}
                  </option>
                  {examFormSections.map((s) => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </select>
              </div>
            </div>
            <div className="grid-3">
              <div className="form-group">
                <label>Subject</label>
                <select
                  value={examForm.subject_id}
                  onChange={(e) => setExamForm((p) => ({ ...p, subject_id: e.target.value }))}
                  disabled={!canPickExamSubject}
                >
                  <option value="">
                    {!examForm.class_id
                      ? "Select class first…"
                      : examFormHasSections && !examForm.section_id
                        ? "Select section first…"
                        : examFormAvailableSubjects.length === 0
                          ? "No subjects with teachers for this section"
                          : "Select subject (optional)…"}
                  </option>
                  {examFormAvailableSubjects.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}{s.code ? ` (${s.code})` : ""}
                      {!s.section_id ? " — class-wide" : ""}
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Exam Date</label>
                <input type="date" required value={examForm.exam_date} onChange={(e) => setExamForm((p) => ({ ...p, exam_date: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Total Marks</label>
                <input type="number" required min="1" value={examForm.total_marks} onChange={(e) => setExamForm((p) => ({ ...p, total_marks: e.target.value }))} placeholder="100" />
              </div>
            </div>
            <p className="text-sm text-muted" style={{ marginTop: 8 }}>
              Only subjects with a teacher assigned for this class and section are listed (see Teacher Assign).
            </p>
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
