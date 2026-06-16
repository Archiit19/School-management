import { useState, useEffect, useCallback, useMemo } from "react";
import { academicApi } from "../api/client";
import PermTabBar from "../components/PermTabBar";
import { usePermTabs } from "../hooks/usePermTabs";
import { useAuth } from "../context/AuthContext";
import {
  classesForTeacher,
  sectionsForAssignedClass,
  subjectsForClassSection,
} from "../utils/teacherClassFilters";

const ASSIGNMENT_TABS = [
  { id: "list", label: "View", perm: "view_assignments" },
  { id: "create", label: "Create Assignment", perm: "create_assignment" },
  { id: "submit", label: "Submit Work", perm: "submit_assignment" },
];

function fmtDate(iso) {
  if (!iso) return "—";
  return iso.split("T")[0];
}

function fmtMarks(marks) {
  if (marks == null) return "—";
  return `${marks}/20`;
}

export default function AssignmentsPage() {
  const { user, hasPerm } = useAuth();
  const isTeacher = user?.role_name === "teacher";
  const canReview = hasPerm("create_assignment");
  const { visibleTabs, tab, setTab } = usePermTabs(ASSIGNMENT_TABS, "list");
  const [assignments, setAssignments] = useState([]);
  const [classes, setClasses] = useState([]);
  const [teacherAssignments, setTeacherAssignments] = useState([]);
  const [query, setQuery] = useState({ class_id: "", subject_id: "", teacher_id: "" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const [form, setForm] = useState({
    class_id: "",
    section_id: "",
    subject_id: "",
    title: "",
    description: "",
    material_url: "",
    due_date: "",
  });
  const [subForm, setSubForm] = useState({ assignment_id: "", student_id: "", content: "", material_url: "" });
  const [viewAssignment, setViewAssignment] = useState(null);
  const [submissions, setSubmissions] = useState([]);
  const [submissionsLoading, setSubmissionsLoading] = useState(false);
  const [reviewingId, setReviewingId] = useState(null);
  const [feedbackDraft, setFeedbackDraft] = useState("");
  const [marksDraft, setMarksDraft] = useState("");

  const load = useCallback(async () => {
    try {
      const q = { ...query };
      if (isTeacher && user?.id && !q.teacher_id) q.teacher_id = user.id;
      setAssignments(await academicApi.getAssignments(q));
    }
    catch (err) { setError(err.message); }
  }, [query, isTeacher, user?.id]);

  useEffect(() => { load(); }, [load]);
  useEffect(() => {
    academicApi.getClasses().then(setClasses).catch(() => {});
    if (isTeacher && user?.id) {
      academicApi
        .getTeacherAssignments({ teacher_user_id: user.id })
        .then(setTeacherAssignments)
        .catch(() => setTeacherAssignments([]));
    }
  }, [isTeacher, user?.id]);

  const flatClasses = classes.map((c) => c.class || c);
  const flatSubjects = classes.flatMap((c) => (c.subjects || []).map((s) => ({ ...s, className: (c.class || c).name })));

  const createClassNode = useMemo(
    () => classes.find((c) => (c.class || c).id === form.class_id),
    [classes, form.class_id],
  );
  const createClassHasSections = (createClassNode?.sections || []).length > 0;
  const createAvailableClasses = useMemo(
    () => classesForTeacher(classes, teacherAssignments, isTeacher),
    [classes, teacherAssignments, isTeacher],
  );
  const createAvailableSections = useMemo(
    () => sectionsForAssignedClass(createClassNode, form.class_id, teacherAssignments, isTeacher),
    [createClassNode, form.class_id, teacherAssignments, isTeacher],
  );
  const createAvailableSubjects = useMemo(
    () => subjectsForClassSection(createClassNode, form.class_id, form.section_id, teacherAssignments, isTeacher),
    [createClassNode, form.class_id, form.section_id, teacherAssignments, isTeacher],
  );
  const canPickCreateSection = form.class_id && createClassHasSections;
  const canPickCreateSubject = form.class_id && (!createClassHasSections || form.section_id);

  function msg(txt) { setSuccess(txt); setError(""); setTimeout(() => setSuccess(""), 3000); }

  async function handleCreate(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try {
      if (createClassHasSections && !form.section_id) {
        setError("Please select a section.");
        setBusy(false);
        return;
      }
      const { section_id: _section, ...rest } = form;
      const payload = { ...rest };
      if (!payload.due_date) delete payload.due_date;
      await academicApi.createAssignment(payload);
      msg("Assignment created.");
      setForm({
        class_id: "",
        section_id: "",
        subject_id: "",
        title: "",
        description: "",
        material_url: "",
        due_date: "",
      });
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

  function classNameFor(id) {
    const c = flatClasses.find((x) => x.id === id);
    return c?.name || id;
  }

  function subjectNameFor(id) {
    const s = flatSubjects.find((x) => x.id === id);
    return s ? `${s.className} — ${s.name}` : id;
  }

  async function openSubmissions(assignment) {
    setViewAssignment(assignment);
    setReviewingId(null);
    setFeedbackDraft("");
    setMarksDraft("");
    setSubmissions([]);
    setSubmissionsLoading(true);
    setError("");
    try {
      setSubmissions(await academicApi.getAssignmentSubmissions(assignment.id));
    } catch (err) {
      setError(err.message);
      setViewAssignment(null);
    } finally {
      setSubmissionsLoading(false);
    }
  }

  function closeSubmissions() {
    setViewAssignment(null);
    setSubmissions([]);
    setReviewingId(null);
    setFeedbackDraft("");
    setMarksDraft("");
  }

  function startReview(sub) {
    setReviewingId(sub.id);
    setFeedbackDraft(sub.teacher_feedback || "");
    setMarksDraft(sub.marks != null ? String(sub.marks) : "");
  }

  async function saveReview(submissionId) {
    setError("");
    const trimmedMarks = marksDraft.trim();
    let marks;
    if (trimmedMarks !== "") {
      marks = Number(trimmedMarks);
      if (!Number.isInteger(marks) || marks < 0 || marks > 20) {
        setError("Marks must be a whole number from 0 to 20.");
        return;
      }
    }
    setBusy(true);
    try {
      const body = { teacher_feedback: feedbackDraft };
      if (marks != null) body.marks = marks;
      await academicApi.reviewSubmission(submissionId, body);
      msg("Evaluation saved.");
      setSubmissions(await academicApi.getAssignmentSubmissions(viewAssignment.id));
      setReviewingId(null);
      setFeedbackDraft("");
      setMarksDraft("");
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <div className="page-header">
        <h1>Assignments & Submissions</h1>
        <p>Flow 7 — Create assignments and record student submissions.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <PermTabBar tabs={visibleTabs} active={tab} onChange={setTab} />

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
            {!isTeacher && (
              <div className="form-group"><label>Teacher ID</label><input placeholder="UUID..." value={query.teacher_id} onChange={(e) => setQuery((q) => ({ ...q, teacher_id: e.target.value }))} /></div>
            )}
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Title</th>
                  <th>Class</th>
                  <th>Subject</th>
                  <th>Due Date</th>
                  {!isTeacher && <th>Teacher ID</th>}
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {assignments.length === 0 && <tr><td colSpan={isTeacher ? 5 : 6} className="empty">No assignments.</td></tr>}
                {assignments.map((a) => (
                  <tr key={a.id}>
                    <td><strong>{a.title}</strong></td>
                    <td>{classNameFor(a.class_id)}</td>
                    <td>{subjectNameFor(a.subject_id)}</td>
                    <td>{fmtDate(a.due_date)}</td>
                    {!isTeacher && <td><span className="mono truncate">{a.teacher_user_id}</span></td>}
                    <td>
                      <button type="button" className="btn btn-ghost btn-sm" onClick={() => openSubmissions(a)}>
                        View
                      </button>
                    </td>
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
          {isTeacher && (
            <p className="text-muted mb-4">
              Choose class, section, then subject — only classes and subjects you are assigned to teach are listed.
            </p>
          )}
          <form onSubmit={handleCreate}>
            <div className="grid-4">
              <div className="form-group">
                <label>Class {isTeacher && <span className="text-muted">(assigned)</span>}</label>
                <select
                  required
                  value={form.class_id}
                  onChange={(e) => setForm((p) => ({ ...p, class_id: e.target.value, section_id: "", subject_id: "" }))}
                >
                  <option value="">Select...</option>
                  {createAvailableClasses.map((c) => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Section {isTeacher && createClassHasSections && <span className="text-muted">(you teach here)</span>}</label>
                <select
                  value={form.section_id}
                  required={createClassHasSections}
                  disabled={!canPickCreateSection}
                  onChange={(e) => setForm((p) => ({ ...p, section_id: e.target.value, subject_id: "" }))}
                >
                  <option value="">
                    {!form.class_id
                      ? "Select class first…"
                      : !createClassHasSections
                        ? "No sections"
                        : createAvailableSections.length === 0
                          ? "No assigned sections"
                          : "Select section…"}
                  </option>
                  {createAvailableSections.map((s) => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Subject {isTeacher && <span className="text-muted">(you teach)</span>}</label>
                <select
                  required
                  value={form.subject_id}
                  disabled={!canPickCreateSubject}
                  onChange={(e) => setForm((p) => ({ ...p, subject_id: e.target.value }))}
                >
                  <option value="">
                    {!canPickCreateSubject
                      ? createClassHasSections ? "Select section first…" : "Select class first…"
                      : createAvailableSubjects.length === 0
                        ? "No subjects for this section"
                        : "Select subject…"}
                  </option>
                  {createAvailableSubjects.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}{s.code ? ` (${s.code})` : ""}
                      {!s.section_id ? " — class-wide" : ""}
                    </option>
                  ))}
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

      {viewAssignment && (
        <div className="modal-overlay" onClick={closeSubmissions}>
          <div className="modal" style={{ maxWidth: 720 }} onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Submissions — {viewAssignment.title}</h3>
              <button type="button" className="modal-close" onClick={closeSubmissions} aria-label="Close">&times;</button>
            </div>
            <p className="text-muted mb-4">
              {classNameFor(viewAssignment.class_id)} · {subjectNameFor(viewAssignment.subject_id)}
              {viewAssignment.due_date && ` · Due ${fmtDate(viewAssignment.due_date)}`}
            </p>
            {viewAssignment.description && (
              <p className="mb-4">{viewAssignment.description}</p>
            )}
            {submissionsLoading ? (
              <p className="text-muted">Loading submissions…</p>
            ) : submissions.length === 0 ? (
              <p className="empty">No submissions yet for this assignment.</p>
            ) : (
              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>Student</th>
                      <th>Submitted</th>
                      <th>Marks</th>
                      <th>Status</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {submissions.map((s) => (
                      <tr key={s.id}>
                        <td>
                          <strong>{s.student_name || "Student"}</strong>
                          {s.student_code && <div className="text-muted">{s.student_code}</div>}
                        </td>
                        <td>{fmtDate(s.created_at)}</td>
                        <td>{fmtMarks(s.marks)}</td>
                        <td>
                          {s.reviewed_at ? (
                            <span className="status status-active">Reviewed</span>
                          ) : (
                            <span className="status status-inactive">Pending</span>
                          )}
                        </td>
                        <td>
                          <button type="button" className="btn btn-ghost btn-sm" onClick={() => startReview(s)}>
                            {canReview ? "Evaluate" : "View"}
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            {reviewingId && (() => {
              const sub = submissions.find((s) => s.id === reviewingId);
              if (!sub) return null;
              return (
                <div className="card" style={{ marginTop: 16 }}>
                  <div className="card-title">
                    {sub.student_name || "Student"}
                    {sub.submitter_name && (
                      <span className="text-muted"> · submitted by {sub.submitter_name}</span>
                    )}
                  </div>
                  <div className="form-group">
                    <label>Student work</label>
                    <div className="review-box">{sub.content || <span className="text-muted">No written content.</span>}</div>
                  </div>
                  {sub.material_url && (
                    <div className="form-group">
                      <label>Attachment</label>
                      <a href={sub.material_url} target="_blank" rel="noreferrer">Open material</a>
                    </div>
                  )}
                  {sub.teacher_feedback && !canReview && (
                    <div className="form-group">
                      <label>Teacher feedback</label>
                      <div className="review-box">{sub.teacher_feedback}</div>
                    </div>
                  )}
                  {sub.marks != null && !canReview && (
                    <div className="form-group">
                      <label>Marks</label>
                      <div className="review-box">{fmtMarks(sub.marks)}</div>
                    </div>
                  )}
                  {canReview && (
                    <>
                      <div className="form-group">
                        <label>Marks (out of 20)</label>
                        <input
                          type="number"
                          min={0}
                          max={20}
                          step={1}
                          value={marksDraft}
                          onChange={(e) => setMarksDraft(e.target.value)}
                          placeholder="0–20"
                        />
                      </div>
                      <div className="form-group">
                        <label>Your feedback / evaluation</label>
                        <textarea
                          rows={4}
                          value={feedbackDraft}
                          onChange={(e) => setFeedbackDraft(e.target.value)}
                          placeholder="Comments for the student or parent…"
                        />
                      </div>
                      <div className="btn-row">
                        <button type="button" className="btn btn-primary" disabled={busy} onClick={() => saveReview(sub.id)}>
                          {busy ? "Saving…" : "Save evaluation"}
                        </button>
                        <button type="button" className="btn btn-ghost" onClick={() => setReviewingId(null)}>Cancel</button>
                      </div>
                    </>
                  )}
                </div>
              );
            })()}
          </div>
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
