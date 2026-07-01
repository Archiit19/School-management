import { useState, useEffect, useCallback, useMemo } from "react";
import { examApi, academicApi, studentApi } from "../api/client";
import PermTabBar from "../components/PermTabBar";
import { usePermTabs } from "../hooks/usePermTabs";
import { useAuth } from "../context/AuthContext";
import { subjectsForClassSection } from "../utils/teacherClassFilters";

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
  const { hasPerm } = useAuth();
  const canManageExams = hasPerm("create_exam");
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
  const [marksEntry, setMarksEntry] = useState({ class_id: "", section_id: "", exam_id: "" });
  const [marksExamList, setMarksExamList] = useState([]);
  const [marksStudents, setMarksStudents] = useState([]);
  const [marksByStudent, setMarksByStudent] = useState({});
  const [marksLoading, setMarksLoading] = useState(false);
  const [marksSaving, setMarksSaving] = useState(false);
  const [publishEntry, setPublishEntry] = useState({ class_id: "", section_id: "", exam_id: "" });
  const [publishExamList, setPublishExamList] = useState([]);
  const [publishStudents, setPublishStudents] = useState([]);
  const [publishMarksByStudent, setPublishMarksByStudent] = useState({});
  const [publishLoading, setPublishLoading] = useState(false);
  const [publishSaving, setPublishSaving] = useState(false);
  const [editingExam, setEditingExam] = useState(null);
  const [editForm, setEditForm] = useState({ class_id: "", section_id: "", subject_id: "", title: "", exam_date: "", total_marks: "" });
  const [editSaving, setEditSaving] = useState(false);

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

  useEffect(() => { if (tab === "exams" || tab === "exam") loadExams(); }, [loadExams, tab]);
  useEffect(() => { if (tab === "results") loadResults(); }, [loadResults, tab]);
  useEffect(() => { academicApi.getClasses().then(setClasses).catch(() => {}); }, []);

  const flatClasses = classes.map((c) => c.class || c);
  const flatSubjects = classes.flatMap((c) => (c.subjects || []).map((s) => ({ ...s, class_id: (c.class || c).id, className: (c.class || c).name })));
  const flatSections = classes.flatMap((c) => (c.sections || []).map((s) => ({ ...s, class_id: (c.class || c).id, className: (c.class || c).name })));

  const examFormClassNode = useMemo(
    () => classes.find((c) => classNodeId(c) === examForm.class_id),
    [classes, examForm.class_id],
  );
  const examFormSections = examFormClassNode?.sections || [];
  const examFormHasSections = examFormSections.length > 0;

  const examFormAvailableSubjects = useMemo(
    () => subjectsForClassSection(examFormClassNode, examForm.class_id, examForm.section_id, [], false),
    [examFormClassNode, examForm.class_id, examForm.section_id],
  );

  const canPickExamSubject = examForm.class_id && (!examFormHasSections || examForm.section_id);

  const editFormClassNode = useMemo(
    () => classes.find((c) => classNodeId(c) === editForm.class_id),
    [classes, editForm.class_id],
  );
  const editFormSections = editFormClassNode?.sections || [];
  const editFormHasSections = editFormSections.length > 0;
  const editFormAvailableSubjects = useMemo(
    () => subjectsForClassSection(editFormClassNode, editForm.class_id, editForm.section_id, [], false),
    [editFormClassNode, editForm.class_id, editForm.section_id],
  );
  const canPickEditSubject = editForm.class_id && (!editFormHasSections || editForm.section_id);

  function examStatusLabel(e) {
    if (e.is_published) return "Published";
    if (e.is_complete) return "Complete";
    return "Draft";
  }

  function openEditExam(exam) {
    setEditingExam(exam);
    setEditForm({
      class_id: exam.class_id || "",
      section_id: exam.section_id || "",
      subject_id: exam.subject_id || "",
      title: exam.title || "",
      exam_date: exam.exam_date?.split("T")[0] || "",
      total_marks: String(exam.total_marks ?? ""),
    });
  }

  function closeEditExam() {
    setEditingExam(null);
    setEditForm({ class_id: "", section_id: "", subject_id: "", title: "", exam_date: "", total_marks: "" });
  }

  function onEditFormClassChange(classId) {
    setEditForm((p) => ({ ...p, class_id: classId, section_id: "", subject_id: "" }));
  }

  function onEditFormSectionChange(sectionId) {
    setEditForm((p) => ({ ...p, section_id: sectionId, subject_id: "" }));
  }

  async function handleUpdateExam(e) {
    e.preventDefault();
    if (!editingExam) return;
    setError("");
    if (!editForm.subject_id) {
      setError("Select a subject.");
      return;
    }
    setEditSaving(true);
    try {
      const body = {
        title: editForm.title,
        class_id: editForm.class_id,
        exam_date: editForm.exam_date,
        total_marks: parseFloat(editForm.total_marks),
        section_id: editForm.section_id || "",
        subject_id: editForm.subject_id,
      };
      await examApi.updateExam(editingExam.id, body);
      msg("Exam updated.");
      closeEditExam();
      loadExams();
    } catch (err) {
      setError(err.message);
    } finally {
      setEditSaving(false);
    }
  }

  async function handleCompleteExam(exam) {
    if (!window.confirm(`Mark "${exam.title}" as complete?`)) return;
    setError("");
    try {
      await examApi.completeExam(exam.id);
      msg("Exam marked complete.");
      loadExams();
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleDeleteExam(exam) {
    if (!window.confirm(`Delete "${exam.title}"? This also removes entered marks.`)) return;
    setError("");
    try {
      await examApi.deleteExam(exam.id);
      msg("Exam deleted.");
      if (editingExam?.id === exam.id) closeEditExam();
      loadExams();
    } catch (err) {
      setError(err.message);
    }
  }

  function onExamFormClassChange(classId) {
    setExamForm((p) => ({ ...p, class_id: classId, section_id: "", subject_id: "" }));
  }

  function onExamFormSectionChange(sectionId) {
    setExamForm((p) => ({ ...p, section_id: sectionId, subject_id: "" }));
  }

  const marksClassNode = useMemo(
    () => classes.find((c) => classNodeId(c) === marksEntry.class_id),
    [classes, marksEntry.class_id],
  );
  const marksSections = marksClassNode?.sections || [];
  const marksHasSections = marksSections.length > 0;
  const canPickMarksExam = marksEntry.class_id && (!marksHasSections || marksEntry.section_id);

  const marksExamsForSelection = useMemo(() => {
    if (!marksEntry.class_id) return [];
    return marksExamList.filter((ex) => {
      if (ex.is_published) return false;
      if (!marksHasSections || !marksEntry.section_id) return true;
      return !ex.section_id || ex.section_id === marksEntry.section_id;
    });
  }, [marksExamList, marksEntry.class_id, marksEntry.section_id, marksHasSections]);

  const selectedMarksExam = useMemo(
    () => marksExamList.find((ex) => ex.id === marksEntry.exam_id),
    [marksExamList, marksEntry.exam_id],
  );

  useEffect(() => {
    if (!marksEntry.class_id) {
      setMarksExamList([]);
      return;
    }
    examApi
      .getExams({ class_id: marksEntry.class_id })
      .then(setMarksExamList)
      .catch(() => setMarksExamList([]));
  }, [marksEntry.class_id]);

  useEffect(() => {
    if (!marksEntry.exam_id || !marksEntry.class_id) {
      setMarksStudents([]);
      setMarksByStudent({});
      return;
    }
    let cancelled = false;
    (async () => {
      setMarksLoading(true);
      setError("");
      try {
        const q = { class_id: marksEntry.class_id, limit: 200, page: 1 };
        if (marksEntry.section_id) q.section_id = marksEntry.section_id;
        const [studRes, existing] = await Promise.all([
          studentApi.list(q),
          examApi.getResults({ exam_id: marksEntry.exam_id }),
        ]);
        if (cancelled) return;
        const students = studRes.students || [];
        setMarksStudents(students);
        const rows = {};
        students.forEach((s) => {
          rows[s.id] = { marks_obtained: "", remarks: "" };
        });
        (existing || []).forEach((r) => {
          if (rows[r.student_id]) {
            rows[r.student_id] = {
              marks_obtained: String(r.marks_obtained ?? ""),
              remarks: "",
            };
          }
        });
        setMarksByStudent(rows);
      } catch (err) {
        if (!cancelled) setError(err.message);
      } finally {
        if (!cancelled) setMarksLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, [marksEntry.exam_id, marksEntry.class_id, marksEntry.section_id]);

  function onMarksClassChange(classId) {
    setMarksEntry({ class_id: classId, section_id: "", exam_id: "" });
    setMarksStudents([]);
    setMarksByStudent({});
  }

  function onMarksSectionChange(sectionId) {
    setMarksEntry((p) => ({ ...p, section_id: sectionId, exam_id: "" }));
    setMarksStudents([]);
    setMarksByStudent({});
  }

  function setStudentMark(studentId, field, value) {
    setMarksByStudent((prev) => ({
      ...prev,
      [studentId]: { ...prev[studentId], [field]: value },
    }));
  }

  const publishClassNode = useMemo(
    () => classes.find((c) => classNodeId(c) === publishEntry.class_id),
    [classes, publishEntry.class_id],
  );
  const publishSections = publishClassNode?.sections || [];
  const publishHasSections = publishSections.length > 0;
  const canPickPublishExam = publishEntry.class_id && (!publishHasSections || publishEntry.section_id);

  const publishExamsForSelection = useMemo(() => {
    if (!publishEntry.class_id) return [];
    return publishExamList.filter((ex) => {
      if (ex.is_published) return false;
      if (!publishHasSections || !publishEntry.section_id) return true;
      return !ex.section_id || ex.section_id === publishEntry.section_id;
    });
  }, [publishExamList, publishEntry.class_id, publishEntry.section_id, publishHasSections]);

  const selectedPublishExam = useMemo(
    () => publishExamList.find((ex) => ex.id === publishEntry.exam_id),
    [publishExamList, publishEntry.exam_id],
  );

  useEffect(() => {
    if (!publishEntry.class_id) {
      setPublishExamList([]);
      return;
    }
    examApi
      .getExams({ class_id: publishEntry.class_id })
      .then(setPublishExamList)
      .catch(() => setPublishExamList([]));
  }, [publishEntry.class_id]);

  useEffect(() => {
    if (!publishEntry.exam_id || !publishEntry.class_id) {
      setPublishStudents([]);
      setPublishMarksByStudent({});
      return;
    }
    let cancelled = false;
    (async () => {
      setPublishLoading(true);
      setError("");
      try {
        const q = { class_id: publishEntry.class_id, limit: 200, page: 1 };
        if (publishEntry.section_id) q.section_id = publishEntry.section_id;
        const [studRes, existing] = await Promise.all([
          studentApi.list(q),
          examApi.getResults({ exam_id: publishEntry.exam_id }),
        ]);
        if (cancelled) return;
        const students = studRes.students || [];
        setPublishStudents(students);
        const marks = {};
        students.forEach((s) => {
          marks[s.id] = null;
        });
        (existing || []).forEach((r) => {
          if (Object.prototype.hasOwnProperty.call(marks, r.student_id)) {
            marks[r.student_id] = r.marks_obtained;
          }
        });
        setPublishMarksByStudent(marks);
      } catch (err) {
        if (!cancelled) setError(err.message);
      } finally {
        if (!cancelled) setPublishLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, [publishEntry.exam_id, publishEntry.class_id, publishEntry.section_id]);

  function onPublishClassChange(classId) {
    setPublishEntry({ class_id: classId, section_id: "", exam_id: "" });
    setPublishStudents([]);
    setPublishMarksByStudent({});
  }

  function onPublishSectionChange(sectionId) {
    setPublishEntry((p) => ({ ...p, section_id: sectionId, exam_id: "" }));
    setPublishStudents([]);
    setPublishMarksByStudent({});
  }

  const publishMarkedCount = useMemo(
    () => Object.values(publishMarksByStudent).filter((m) => m != null && m !== "").length,
    [publishMarksByStudent],
  );

  async function handleSaveAllMarks(e) {
    e.preventDefault();
    setError("");
    if (!selectedMarksExam) {
      setError("Select an exam.");
      return;
    }
    if (selectedMarksExam.is_published) {
      setError("Cannot enter marks after results are published.");
      return;
    }
    const max = selectedMarksExam.total_marks;
    const toSave = [];
    for (const st of marksStudents) {
      const row = marksByStudent[st.id];
      if (!row || String(row.marks_obtained).trim() === "") continue;
      const m = parseFloat(row.marks_obtained);
      if (Number.isNaN(m) || m < 0 || m > max) {
        setError(`Marks for ${st.first_name} ${st.last_name} must be between 0 and ${max}.`);
        return;
      }
      toSave.push({
        student_id: st.id,
        marks_obtained: m,
        remarks: row.remarks || "",
      });
    }
    if (toSave.length === 0) {
      setError("Enter marks for at least one student (leave blank to skip).");
      return;
    }
    setMarksSaving(true);
    try {
      for (const row of toSave) {
        await examApi.enterMarks({
          exam_id: marksEntry.exam_id,
          student_id: row.student_id,
          marks_obtained: row.marks_obtained,
          remarks: row.remarks,
        });
      }
      msg(`Saved marks for ${toSave.length} student(s).`);
    } catch (err) {
      setError(err.message);
    } finally {
      setMarksSaving(false);
    }
  }

  function msg(txt) { setSuccess(txt); setError(""); setTimeout(() => setSuccess(""), 3000); }

  function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => msg("Copied to clipboard!")).catch(() => setError("Failed to copy"));
  }

  async function handleCreateExam(e) {
    e.preventDefault();
    setError("");
    if (!examForm.subject_id) {
      setError("Select a subject.");
      return;
    }
    setBusy(true);
    try {
      const payload = { ...examForm, total_marks: parseFloat(examForm.total_marks) };
      if (!payload.section_id) delete payload.section_id;
      const created = await examApi.createExam(payload);
      msg(`Exam created! ID: ${created.id}`);
      setExamForm({ class_id: "", section_id: "", subject_id: "", title: "", exam_date: "", total_marks: "" });
      loadExams();
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  const classNameById = new Map(classes.map((c) => { const cls = c.class || c; return [cls.id, cls.name]; }));
  const sectionNameById = new Map(classes.flatMap((c) => (c.sections || []).map((s) => [s.id, s.name])));
  const subjectNameById = new Map(classes.flatMap((c) => (c.subjects || []).map((s) => [s.id, s.name])));

  async function handlePublish(e) {
    e.preventDefault();
    setError("");
    if (!selectedPublishExam) {
      setError("Select an exam.");
      return;
    }
    if (selectedPublishExam.is_published) {
      setError("This exam is already published.");
      return;
    }
    if (publishMarkedCount === 0) {
      setError("Enter marks for at least one student before publishing.");
      return;
    }
    setPublishSaving(true);
    try {
      await examApi.publish({ exam_id: publishEntry.exam_id });
      msg("Results published. Students can now view them in My Portal.");
      setPublishEntry((p) => ({ ...p, exam_id: "" }));
      setPublishStudents([]);
      setPublishMarksByStudent({});
      if (publishEntry.class_id) {
        const list = await examApi.getExams({ class_id: publishEntry.class_id });
        setPublishExamList(list);
      }
      loadExams();
    } catch (err) {
      setError(err.message);
    } finally {
      setPublishSaving(false);
    }
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
              <thead><tr><th>Title</th><th>Class</th><th>Section</th><th>Subject</th><th>Date</th><th>Total</th><th>Status</th>{canManageExams && <th>Actions</th>}<th>Exam ID</th></tr></thead>
              <tbody>
                {(!exams || exams.length === 0) && <tr><td colSpan={canManageExams ? 9 : 8} className="empty">No exams found.</td></tr>}
                {(exams || []).map((e) => (
                  <tr key={e.id}>
                    <td><strong>{e.title}</strong></td>
                    <td>{classNameById.get(e.class_id) || <span className="mono truncate">{e.class_id}</span>}</td>
                    <td>{e.section_id ? (sectionNameById.get(e.section_id) || <span className="mono truncate">{e.section_id}</span>) : "—"}</td>
                    <td>{e.subject_id ? (subjectNameById.get(e.subject_id) || <span className="mono truncate">{e.subject_id}</span>) : "—"}</td>
                    <td>{e.exam_date?.split("T")[0]}</td>
                    <td>{e.total_marks}</td>
                    <td>
                      <span className={`status ${e.is_published ? "status-active" : e.is_complete ? "status-late" : "status-inactive"}`}>
                        {examStatusLabel(e)}
                      </span>
                    </td>
                    {canManageExams && (
                      <td>
                        {!e.is_published && (
                          <div className="btn-row" style={{ gap: 4 }}>
                            <button type="button" className="btn btn-secondary btn-sm" onClick={() => openEditExam(e)}>Edit</button>
                            {!e.is_complete && (
                              <button type="button" className="btn btn-secondary btn-sm" onClick={() => handleCompleteExam(e)}>Complete</button>
                            )}
                            <button type="button" className="btn btn-danger btn-sm" onClick={() => handleDeleteExam(e)}>Delete</button>
                          </div>
                        )}
                        {e.is_published && <span className="text-muted text-sm">—</span>}
                      </td>
                    )}
                    <td><span className="mono truncate">{e.id}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {editingExam && (
        <div className="card" style={{ marginTop: 16 }}>
          <div className="card-title">Edit Exam</div>
          <form onSubmit={handleUpdateExam}>
            <div className="grid-3">
              <div className="form-group">
                <label>Title</label>
                <input required value={editForm.title} onChange={(ev) => setEditForm((p) => ({ ...p, title: ev.target.value }))} />
              </div>
              <div className="form-group">
                <label>Class</label>
                <select required value={editForm.class_id} onChange={(ev) => onEditFormClassChange(ev.target.value)}>
                  <option value="">Select class…</option>
                  {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Section</label>
                <select
                  required={editFormHasSections}
                  disabled={!editForm.class_id}
                  value={editForm.section_id}
                  onChange={(ev) => onEditFormSectionChange(ev.target.value)}
                >
                  <option value="">
                    {!editForm.class_id ? "Select class first…" : editFormHasSections ? "Select section…" : "No sections (class-wide)"}
                  </option>
                  {editFormSections.map((s) => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </select>
              </div>
            </div>
            <div className="grid-3">
              <div className="form-group">
                <label>Subject</label>
                <select
                  required
                  value={editForm.subject_id}
                  onChange={(ev) => setEditForm((p) => ({ ...p, subject_id: ev.target.value }))}
                  disabled={!canPickEditSubject}
                >
                  <option value="">
                    {!editForm.class_id
                      ? "Select class first…"
                      : editFormHasSections && !editForm.section_id
                        ? "Select section first…"
                        : editFormAvailableSubjects.length === 0
                          ? "No subjects for this section"
                          : "Select subject…"}
                  </option>
                  {editFormAvailableSubjects.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}{s.code ? ` (${s.code})` : ""}
                      {!s.section_id ? " — class-wide" : ""}
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Exam Date</label>
                <input type="date" required value={editForm.exam_date} onChange={(ev) => setEditForm((p) => ({ ...p, exam_date: ev.target.value }))} />
              </div>
              <div className="form-group">
                <label>Total Marks</label>
                <input type="number" required min="1" value={editForm.total_marks} onChange={(ev) => setEditForm((p) => ({ ...p, total_marks: ev.target.value }))} />
              </div>
            </div>
            <div className="btn-row">
              <button type="submit" className="btn btn-primary" disabled={editSaving}>{editSaving ? "Saving…" : "Save changes"}</button>
              <button type="button" className="btn btn-secondary" onClick={closeEditExam}>Cancel</button>
            </div>
          </form>
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
                  required
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
                          ? "No subjects for this section"
                          : "Select subject…"}
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
              All subjects configured for this class and section are listed.
            </p>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Create Exam</button></div>
          </form>
        </div>
      )}

      {tab === "marks" && (
        <>
          <div className="card">
            <div className="card-title">Enter Marks <span className="badge badge-post">POST /marks</span></div>
            <form onSubmit={handleSaveAllMarks}>
              <div className="grid-3 mb-4">
                <div className="form-group">
                  <label>Class</label>
                  <select required value={marksEntry.class_id} onChange={(e) => onMarksClassChange(e.target.value)}>
                    <option value="">Select class…</option>
                    {flatClasses.map((c) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </select>
                </div>
                <div className="form-group">
                  <label>Section</label>
                  <select
                    required={marksHasSections}
                    disabled={!marksEntry.class_id}
                    value={marksEntry.section_id}
                    onChange={(e) => onMarksSectionChange(e.target.value)}
                  >
                    <option value="">
                      {!marksEntry.class_id
                        ? "Select class first…"
                        : marksHasSections
                          ? "Select section…"
                          : "No sections"}
                    </option>
                    {marksSections.map((s) => (
                      <option key={s.id} value={s.id}>{s.name}</option>
                    ))}
                  </select>
                </div>
                <div className="form-group">
                  <label>Exam</label>
                  <select
                    required
                    disabled={!canPickMarksExam}
                    value={marksEntry.exam_id}
                    onChange={(e) => setMarksEntry((p) => ({ ...p, exam_id: e.target.value }))}
                  >
                    <option value="">
                      {!marksEntry.class_id
                        ? "Select class first…"
                        : marksHasSections && !marksEntry.section_id
                          ? "Select section first…"
                          : marksExamsForSelection.length === 0
                            ? "No draft exams for this class/section"
                            : "Select exam…"}
                    </option>
                    {marksExamsForSelection.map((ex) => (
                      <option key={ex.id} value={ex.id}>
                        {ex.title} — {ex.exam_date?.split("T")[0]} (max {ex.total_marks})
                      </option>
                    ))}
                  </select>
                </div>
              </div>
              {selectedMarksExam && (
                <p className="text-sm text-muted" style={{ marginBottom: 12 }}>
                  Maximum marks: <strong>{selectedMarksExam.total_marks}</strong>. Leave blank to skip a student. Existing marks are pre-filled.
                </p>
              )}
              {marksLoading && <div className="empty">Loading students…</div>}
              {!marksLoading && marksEntry.exam_id && marksStudents.length === 0 && (
                <div className="empty">No students in this class/section.</div>
              )}
              {!marksLoading && marksStudents.length > 0 && (
                <div className="table-wrap">
                  <table>
                    <thead>
                      <tr>
                        <th>Student</th>
                        <th>Code</th>
                        <th>Marks obtained</th>
                        <th>Remarks</th>
                      </tr>
                    </thead>
                    <tbody>
                      {marksStudents.map((st) => {
                        const row = marksByStudent[st.id] || { marks_obtained: "", remarks: "" };
                        const max = selectedMarksExam?.total_marks ?? 100;
                        const m = parseFloat(row.marks_obtained);
                        const invalid = row.marks_obtained !== "" && (Number.isNaN(m) || m < 0 || m > max);
                        return (
                          <tr key={st.id}>
                            <td><strong>{st.first_name} {st.last_name}</strong></td>
                            <td>{st.student_code || "—"}</td>
                            <td>
                              <input
                                type="number"
                                min={0}
                                max={max}
                                step="0.01"
                                value={row.marks_obtained}
                                onChange={(e) => setStudentMark(st.id, "marks_obtained", e.target.value)}
                                placeholder={`0–${max}`}
                                style={invalid ? { borderColor: "var(--clr-danger)" } : undefined}
                              />
                              {invalid && (
                                <div className="text-sm" style={{ color: "var(--clr-danger)", marginTop: 4 }}>
                                  Must be 0–{max}
                                </div>
                              )}
                            </td>
                            <td>
                              <input
                                value={row.remarks}
                                onChange={(e) => setStudentMark(st.id, "remarks", e.target.value)}
                                placeholder="Optional"
                              />
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              )}
              <div className="btn-row" style={{ marginTop: 16 }}>
                <button
                  type="submit"
                  className="btn btn-primary"
                  disabled={marksSaving || marksLoading || !marksEntry.exam_id || marksStudents.length === 0}
                >
                  {marksSaving ? "Saving…" : "Save all marks"}
                </button>
              </div>
            </form>
          </div>
        </>
      )}

      {tab === "publish" && (
        <div className="card">
          <div className="card-title">Publish Results <span className="badge badge-post">POST /results/publish</span></div>
          <form onSubmit={handlePublish}>
            <div className="grid-3 mb-4">
              <div className="form-group">
                <label>Class</label>
                <select required value={publishEntry.class_id} onChange={(e) => onPublishClassChange(e.target.value)}>
                  <option value="">Select class…</option>
                  {flatClasses.map((c) => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Section</label>
                <select
                  required={publishHasSections}
                  disabled={!publishEntry.class_id}
                  value={publishEntry.section_id}
                  onChange={(e) => onPublishSectionChange(e.target.value)}
                >
                  <option value="">
                    {!publishEntry.class_id
                      ? "Select class first…"
                      : publishHasSections
                        ? "Select section…"
                        : "No sections"}
                  </option>
                  {publishSections.map((s) => (
                    <option key={s.id} value={s.id}>{s.name}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Exam</label>
                <select
                  required
                  disabled={!canPickPublishExam}
                  value={publishEntry.exam_id}
                  onChange={(e) => setPublishEntry((p) => ({ ...p, exam_id: e.target.value }))}
                >
                  <option value="">
                    {!publishEntry.class_id
                      ? "Select class first…"
                      : publishHasSections && !publishEntry.section_id
                        ? "Select section first…"
                        : publishExamsForSelection.length === 0
                          ? "No draft exams for this class/section"
                          : "Select exam…"}
                  </option>
                  {publishExamsForSelection.map((ex) => (
                    <option key={ex.id} value={ex.id}>
                      {ex.title} — {ex.exam_date?.split("T")[0]} (max {ex.total_marks})
                    </option>
                  ))}
                </select>
              </div>
            </div>
            {selectedPublishExam && (
              <p className="text-sm text-muted" style={{ marginBottom: 12 }}>
                Review marks below, then publish to make results visible to students and parents.
                {publishMarkedCount > 0 && (
                  <> <strong>{publishMarkedCount}</strong> of {publishStudents.length} student(s) have marks entered.</>
                )}
              </p>
            )}
            {publishLoading && <div className="empty">Loading preview…</div>}
            {!publishLoading && publishEntry.exam_id && publishStudents.length === 0 && (
              <div className="empty">No students in this class/section.</div>
            )}
            {!publishLoading && publishStudents.length > 0 && (
              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>Student</th>
                      <th>Code</th>
                      <th>Marks</th>
                      <th>Out of</th>
                    </tr>
                  </thead>
                  <tbody>
                    {publishStudents.map((st) => {
                      const m = publishMarksByStudent[st.id];
                      const hasMark = m != null && m !== "";
                      const max = selectedPublishExam?.total_marks;
                      return (
                        <tr key={st.id}>
                          <td><strong>{st.first_name} {st.last_name}</strong></td>
                          <td>{st.student_code || "—"}</td>
                          <td>
                            {hasMark ? (
                              <strong>{m}</strong>
                            ) : (
                              <span className="text-muted">Not entered</span>
                            )}
                          </td>
                          <td>{hasMark ? max : "—"}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
            <div className="btn-row" style={{ marginTop: 16 }}>
              <button
                type="submit"
                className="btn btn-primary"
                disabled={
                  publishSaving
                  || publishLoading
                  || !publishEntry.exam_id
                  || publishMarkedCount === 0
                }
              >
                {publishSaving ? "Publishing…" : "Publish results"}
              </button>
            </div>
          </form>
        </div>
      )}
    </>
  );
}
