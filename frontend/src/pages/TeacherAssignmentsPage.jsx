import { useState, useEffect, useCallback, useMemo } from "react";
import { academicApi, rolesApi, userMgmtApi } from "../api/client";
import PermGate from "../components/PermGate";
import { useAuth } from "../context/AuthContext";

function classNodeId(node) {
  return (node.class || node).id;
}

export default function TeacherAssignmentsPage() {
  const { hasPerm } = useAuth();
  const [assignments, setAssignments] = useState([]);
  const [allAssignments, setAllAssignments] = useState([]);
  const [classes, setClasses] = useState([]);
  const [teachers, setTeachers] = useState([]);
  const [teachersLoading, setTeachersLoading] = useState(false);
  const [teachersError, setTeachersError] = useState("");
  const [query, setQuery] = useState({ teacher_user_id: "", class_id: "", section_id: "", subject_id: "" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const [form, setForm] = useState({
    teacher_user_id: "",
    class_id: "",
    section_id: "",
    subject_id: "",
  });

  const subjectById = useMemo(() => {
    const map = new Map();
    classes.forEach((node) => {
      (node.subjects || []).forEach((s) => map.set(s.id, s));
    });
    return map;
  }, [classes]);

  const sectionById = useMemo(() => {
    const map = new Map();
    classes.forEach((node) => {
      (node.sections || []).forEach((s) => map.set(s.id, s));
    });
    return map;
  }, [classes]);

  const loadAllAssignments = useCallback(async () => {
    try {
      setAllAssignments(await academicApi.getTeacherAssignments({}));
    } catch {
      setAllAssignments([]);
    }
  }, []);

  const load = useCallback(async () => {
    try {
      const q = { ...query };
      if (!q.section_id) delete q.section_id;
      const rows = await academicApi.getTeacherAssignments(q);
      if (query.section_id) {
        const sectionId = query.section_id;
        setAssignments(rows.filter((a) => {
          const sub = subjectById.get(a.subject_id);
          if (!sub) return false;
          return !sub.section_id || sub.section_id === sectionId;
        }));
      } else {
        setAssignments(rows);
      }
    } catch (err) {
      setError(err.message);
    }
  }, [query, subjectById]);

  useEffect(() => { loadAllAssignments(); }, [loadAllAssignments]);
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
  const classMap = Object.fromEntries(flatClasses.map((c) => [c.id, c.name]));
  const teacherMap = useMemo(
    () => Object.fromEntries(teachers.map((t) => [t.id, t])),
    [teachers],
  );

  const formClassNode = useMemo(
    () => classes.find((c) => classNodeId(c) === form.class_id),
    [classes, form.class_id],
  );
  const formSections = formClassNode?.sections || [];
  const formHasSections = formSections.length > 0;

  const assignedSubjectIdsForClass = useMemo(() => {
    const ids = new Set();
    allAssignments.forEach((a) => {
      if (a.class_id === form.class_id) ids.add(a.subject_id);
    });
    return ids;
  }, [allAssignments, form.class_id]);

  const availableSubjects = useMemo(() => {
    if (!form.class_id || !formClassNode) return [];
    const subjects = formClassNode.subjects || [];
    return subjects.filter((s) => {
      if (assignedSubjectIdsForClass.has(s.id)) return false;
      if (formHasSections) {
        if (!form.section_id) return false;
        if (!s.section_id) return true;
        return s.section_id === form.section_id;
      }
      return !s.section_id;
    });
  }, [form.class_id, form.section_id, formClassNode, formHasSections, assignedSubjectIdsForClass]);

  const queryClassNode = useMemo(
    () => classes.find((c) => classNodeId(c) === query.class_id),
    [classes, query.class_id],
  );
  const querySections = queryClassNode?.sections || [];

  const querySubjects = useMemo(() => {
    if (!query.class_id || !queryClassNode) return [];
    const subjects = queryClassNode.subjects || [];
    if (!query.section_id) return subjects;
    return subjects.filter((s) => !s.section_id || s.section_id === query.section_id);
  }, [query.class_id, query.section_id, queryClassNode]);

  function msg(txt) {
    setSuccess(txt);
    setError("");
    setTimeout(() => setSuccess(""), 3000);
  }

  function onFormClassChange(classId) {
    setForm((p) => ({ ...p, class_id: classId, section_id: "", subject_id: "" }));
  }

  function onFormSectionChange(sectionId) {
    setForm((p) => ({ ...p, section_id: sectionId, subject_id: "" }));
  }

  async function handleCreate(e) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await academicApi.createTeacherAssignment({
        teacher_user_id: form.teacher_user_id,
        class_id: form.class_id,
        subject_id: form.subject_id,
      });
      msg("Teacher assigned.");
      setForm((p) => ({ ...p, subject_id: "" }));
      await loadAllAssignments();
      load();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  const canPickSubject = form.class_id && (!formHasSections || form.section_id);
  const subjectSelectDisabled = !canPickSubject;

  return (
    <>
      <div className="page-header">
        <h1>Teacher Assignments</h1>
        <p>
          Assign one teacher per subject: choose class, then section, then a subject that does not already have a teacher.
        </p>
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
          <div className="grid-4">
            <div className="form-group">
              <label>Teacher</label>
              {hasPerm("view_users") && teachers.length > 0 ? (
                <select
                  name="teacher_user_id"
                  required
                  value={form.teacher_user_id}
                  onChange={(e) => setForm((p) => ({ ...p, teacher_user_id: e.target.value }))}
                >
                  <option value="">Select teacher…</option>
                  {teachers.filter((t) => t.is_active !== false).map((t) => (
                    <option key={t.id} value={t.id}>{t.name} — {t.email}</option>
                  ))}
                </select>
              ) : (
                <input
                  name="teacher_user_id"
                  required
                  value={form.teacher_user_id}
                  onChange={(e) => setForm((p) => ({ ...p, teacher_user_id: e.target.value }))}
                  placeholder="UUID of teacher user"
                />
              )}
            </div>
            <div className="form-group">
              <label>Class</label>
              <select
                required
                value={form.class_id}
                onChange={(e) => onFormClassChange(e.target.value)}
              >
                <option value="">Select class…</option>
                {flatClasses.map((c) => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>Section</label>
              <select
                required={formHasSections}
                disabled={!form.class_id}
                value={form.section_id}
                onChange={(e) => onFormSectionChange(e.target.value)}
              >
                <option value="">
                  {!form.class_id
                    ? "Select class first…"
                    : formHasSections
                      ? "Select section…"
                      : "No sections (class-wide subjects)"}
                </option>
                {formSections.map((s) => (
                  <option key={s.id} value={s.id}>{s.name}</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>Subject</label>
              <select
                name="subject_id"
                required
                disabled={subjectSelectDisabled}
                value={form.subject_id}
                onChange={(e) => setForm((p) => ({ ...p, subject_id: e.target.value }))}
              >
                <option value="">
                  {!form.class_id
                    ? "Select class first…"
                    : formHasSections && !form.section_id
                      ? "Select section first…"
                      : availableSubjects.length === 0
                        ? "No unassigned subjects"
                        : "Select subject…"}
                </option>
                {availableSubjects.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}{s.code ? ` (${s.code})` : ""}
                    {!s.section_id ? " — class-wide" : ""}
                  </option>
                ))}
              </select>
            </div>
          </div>
          <p className="text-sm text-muted" style={{ marginTop: 8 }}>
            Only subjects without a teacher for this class are listed. Class-wide subjects appear for every section.
          </p>
          <div className="btn-row">
            <button className="btn btn-primary" disabled={busy || !form.subject_id}>Assign</button>
          </div>
        </form>
      </div>

      <div className="card">
        <div className="card-title">Class assignments <span className="badge badge-get">GET /teacher-assignments</span></div>
        <div className="grid-4 mb-4">
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
              <input
                placeholder="Filter by teacher UUID…"
                value={query.teacher_user_id}
                onChange={(e) => setQuery((q) => ({ ...q, teacher_user_id: e.target.value }))}
              />
            )}
          </div>
          <div className="form-group">
            <label>Class</label>
            <select
              value={query.class_id}
              onChange={(e) => setQuery((q) => ({ ...q, class_id: e.target.value, section_id: "", subject_id: "" }))}
            >
              <option value="">All</option>
              {flatClasses.map((c) => (
                <option key={c.id} value={c.id}>{c.name}</option>
              ))}
            </select>
          </div>
          <div className="form-group">
            <label>Section</label>
            <select
              value={query.section_id}
              disabled={!query.class_id}
              onChange={(e) => setQuery((q) => ({ ...q, section_id: e.target.value, subject_id: "" }))}
            >
              <option value="">{query.class_id ? "All sections" : "Select class first…"}</option>
              {querySections.map((s) => (
                <option key={s.id} value={s.id}>{s.name}</option>
              ))}
            </select>
          </div>
          <div className="form-group">
            <label>Subject</label>
            <select
              value={query.subject_id}
              disabled={!query.class_id}
              onChange={(e) => setQuery((q) => ({ ...q, subject_id: e.target.value }))}
            >
              <option value="">{query.class_id ? "All subjects" : "Select class first…"}</option>
              {querySubjects.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}{s.section_id ? ` (${sectionById.get(s.section_id)?.name || "section"})` : " (class-wide)"}
                </option>
              ))}
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Teacher</th><th>Class</th><th>Section</th><th>Subject</th><th>Assignment ID</th></tr></thead>
            <tbody>
              {assignments.length === 0 && <tr><td colSpan={5} className="empty">No assignments found.</td></tr>}
              {assignments.map((a) => {
                const sub = subjectById.get(a.subject_id);
                const sectionLabel = sub?.section_id
                  ? (sectionById.get(sub.section_id)?.name || "—")
                  : "Class-wide";
                return (
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
                    <td>{sectionLabel}</td>
                    <td>{sub?.name || a.subject_id}{sub?.code ? ` (${sub.code})` : ""}</td>
                    <td><span className="mono truncate">{a.id}</span></td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
