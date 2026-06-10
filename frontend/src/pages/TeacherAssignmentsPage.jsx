import { useState, useEffect, useCallback, useMemo } from "react";
import { academicApi, rolesApi, userMgmtApi } from "../api/client";
import PermGate from "../components/PermGate";
import { useAuth } from "../context/AuthContext";

function classNodeId(node) {
  return (node.class || node).id;
}

function TeacherSelectField({
  label,
  value,
  onChange,
  teachers,
  lockedTeacher,
  allowAll = false,
  required = false,
  name,
  placeholder = "Select teacher…",
}) {
  const locked = lockedTeacher;
  const options = locked ? [locked] : teachers;

  return (
    <div className="form-group">
      <label>{label}</label>
      <select
        name={name}
        required={required && !locked}
        disabled={!!locked}
        value={locked ? locked.id : value}
        onChange={onChange}
      >
        {!locked && allowAll && <option value="">All teachers</option>}
        {!locked && !allowAll && <option value="">{placeholder}</option>}
        {options.map((t) => (
          <option key={t.id} value={t.id}>
            {t.name}{t.email ? ` — ${t.email}` : ""}
          </option>
        ))}
      </select>
    </div>
  );
}

export default function TeacherAssignmentsPage() {
  const { user, hasPerm } = useAuth();
  const canAssignTeacher = hasPerm("assign_teacher");
  const canListTeachers = hasPerm("view_users");
  const isViewOnly = !canAssignTeacher;
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
  const [editingAssignment, setEditingAssignment] = useState(null);
  const [editForm, setEditForm] = useState({
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
    if (!user?.id || canAssignTeacher) return;
    setQuery((q) => (q.teacher_user_id ? q : { ...q, teacher_user_id: user.id }));
  }, [user?.id, canAssignTeacher]);

  useEffect(() => {
    if (!canListTeachers) {
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
  }, [canListTeachers]);

  const flatClasses = classes.map((c) => c.class || c);
  const classMap = Object.fromEntries(flatClasses.map((c) => [c.id, c.name]));

  const selfTeacher = useMemo(() => {
    if (!user?.id) return null;
    return {
      id: user.id,
      name: user.name,
      email: user.email,
      is_active: user.is_active !== false,
    };
  }, [user]);

  const teachersFromAssignments = useMemo(() => {
    const byId = new Map();
    if (selfTeacher) byId.set(selfTeacher.id, selfTeacher);
    teachers.forEach((t) => byId.set(t.id, t));
    allAssignments.forEach((a) => {
      const id = a.teacher_user_id;
      if (!byId.has(id)) {
        byId.set(id, { id, name: `Teacher ${String(id).slice(0, 8)}…`, email: "", is_active: true });
      }
    });
    return Array.from(byId.values());
  }, [allAssignments, selfTeacher, teachers]);

  const assignFormTeachers = canListTeachers
    ? teachers.filter((t) => t.is_active !== false)
    : teachersFromAssignments.filter((t) => t.is_active !== false);

  const lockedAssignTeacher = !canListTeachers && assignFormTeachers.length === 1 ? assignFormTeachers[0] : null;

  useEffect(() => {
    if (!lockedAssignTeacher) return;
    setForm((f) => (f.teacher_user_id === lockedAssignTeacher.id ? f : { ...f, teacher_user_id: lockedAssignTeacher.id }));
  }, [lockedAssignTeacher]);

  const teacherRows = canListTeachers ? teachers : isViewOnly && selfTeacher ? [selfTeacher] : teachersFromAssignments;

  const teacherMap = useMemo(() => {
    const map = Object.fromEntries(teachersFromAssignments.map((t) => [t.id, t]));
    return map;
  }, [teachersFromAssignments]);

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

  const assignedSubjectIdsForEditClass = useMemo(() => {
    const ids = new Set();
    allAssignments.forEach((a) => {
      if (a.class_id === editForm.class_id) ids.add(a.subject_id);
    });
    return ids;
  }, [allAssignments, editForm.class_id]);

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

  const editClassNode = useMemo(
    () => classes.find((c) => classNodeId(c) === editForm.class_id),
    [classes, editForm.class_id],
  );
  const editSections = editClassNode?.sections || [];
  const editHasSections = editSections.length > 0;

  const editAvailableSubjects = useMemo(() => {
    if (!editForm.class_id || !editClassNode) return [];
    const subjects = editClassNode.subjects || [];
    const currentSubjectId = editingAssignment?.subject_id;
    return subjects.filter((s) => {
      if (assignedSubjectIdsForEditClass.has(s.id) && s.id !== currentSubjectId) return false;
      if (editHasSections) {
        if (!editForm.section_id) return false;
        if (!s.section_id) return true;
        return s.section_id === editForm.section_id;
      }
      return !s.section_id;
    });
  }, [
    editForm.class_id,
    editForm.section_id,
    editClassNode,
    editHasSections,
    assignedSubjectIdsForEditClass,
    editingAssignment?.subject_id,
  ]);

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
    if (!canAssignTeacher) {
      setError("You do not have permission to assign teachers.");
      return;
    }
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

  function openEdit(assignment) {
    const sub = subjectById.get(assignment.subject_id);
    setEditingAssignment(assignment);
    setEditForm({
      teacher_user_id: assignment.teacher_user_id,
      class_id: assignment.class_id,
      section_id: sub?.section_id || "",
      subject_id: assignment.subject_id,
    });
    setError("");
  }

  function closeEdit() {
    setEditingAssignment(null);
  }

  function onEditClassChange(classId) {
    setEditForm((p) => ({ ...p, class_id: classId, section_id: "", subject_id: "" }));
  }

  function onEditSectionChange(sectionId) {
    setEditForm((p) => ({ ...p, section_id: sectionId, subject_id: "" }));
  }

  async function handleUpdate(e) {
    e.preventDefault();
    if (!editingAssignment || !canAssignTeacher) return;
    setError("");
    setBusy(true);
    try {
      await academicApi.updateTeacherAssignment(editingAssignment.id, {
        teacher_user_id: editForm.teacher_user_id,
        class_id: editForm.class_id,
        subject_id: editForm.subject_id,
      });
      msg("Assignment updated.");
      closeEdit();
      await loadAllAssignments();
      load();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  async function handleRemove(assignment) {
    if (!canAssignTeacher) return;
    const sub = subjectById.get(assignment.subject_id);
    const teacherName = teacherMap[assignment.teacher_user_id]?.name || "this teacher";
    const subjectName = sub?.name || "this subject";
    if (!confirm(`Remove ${teacherName} from ${subjectName}?`)) return;
    setError("");
    setBusy(true);
    try {
      await academicApi.deleteTeacherAssignment(assignment.id);
      msg("Assignment removed.");
      await loadAllAssignments();
      load();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  const canPickSubject = form.class_id && (!formHasSections || form.section_id);
  const canPickEditSubject = editForm.class_id && (!editHasSections || editForm.section_id);
  const subjectSelectDisabled = !canPickSubject;

  return (
    <>
      <div className="page-header">
        <h1>Teacher Assignments</h1>
        <p>
          {canAssignTeacher
            ? "Assign one teacher per subject: choose class, then section, then a subject that does not already have a teacher."
            : "View your teaching assignments by class, section, and subject."}
        </p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="card">
        <div className="card-title">
          School teachers <span className="badge badge-get">GET /users?role_id=…</span>
        </div>
        {canListTeachers && teachersError && <div className="alert alert-error">{teachersError}</div>}
        <div className="table-wrap">
          <table>
            <thead><tr><th>Name</th><th>Email</th><th>User ID</th><th>Status</th></tr></thead>
            <tbody>
              {canListTeachers && teachersLoading && <tr><td colSpan={4} className="empty">Loading teachers…</td></tr>}
              {!teachersLoading && teacherRows.length === 0 && (
                <tr><td colSpan={4} className="empty">No teacher accounts found.</td></tr>
              )}
              {!teachersLoading && teacherRows.map((t) => (
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

      <div className="card">
        <div className="card-title">Assign Teacher <span className="badge badge-post">POST /teacher-assignments</span></div>
        {canAssignTeacher ? (
        <form onSubmit={handleCreate}>
          <div className="grid-4">
            <TeacherSelectField
              label="Teacher"
              name="teacher_user_id"
              required
              value={form.teacher_user_id}
              teachers={assignFormTeachers}
              lockedTeacher={lockedAssignTeacher}
              onChange={(e) => setForm((p) => ({ ...p, teacher_user_id: e.target.value }))}
            />
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
        ) : (
          <p className="text-muted" style={{ margin: 0 }}>
            You do not have permission to assign teachers. Use the filters below to view your classes and subjects.
          </p>
        )}
      </div>

      <div className="card">
        <div className="card-title">
          {isViewOnly ? "My teaching assignments" : "Class assignments"}{" "}
          <span className="badge badge-get">GET /teacher-assignments</span>
        </div>
        <div className="grid-4 mb-4">
          <TeacherSelectField
            label="Teacher"
            allowAll={canAssignTeacher}
            value={query.teacher_user_id}
            teachers={canListTeachers ? teachers : teachersFromAssignments}
            lockedTeacher={isViewOnly ? selfTeacher : null}
            onChange={(e) => setQuery((q) => ({ ...q, teacher_user_id: e.target.value }))}
          />
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
            <thead><tr><th>Teacher</th><th>Class</th><th>Section</th><th>Subject</th><th>Actions</th></tr></thead>
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
                    <td>
                      <PermGate perm="assign_teacher">
                        <div className="btn-row" style={{ flexWrap: "nowrap" }}>
                          <button type="button" className="btn btn-ghost btn-sm" onClick={() => openEdit(a)}>Edit</button>
                          <button type="button" className="btn btn-danger btn-sm" onClick={() => handleRemove(a)} disabled={busy}>Remove</button>
                        </div>
                      </PermGate>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      {editingAssignment && (
        <div className="modal-overlay" onClick={closeEdit}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Edit teacher assignment</h3>
              <button type="button" className="modal-close" onClick={closeEdit} aria-label="Close">&times;</button>
            </div>
            <form onSubmit={handleUpdate}>
              <TeacherSelectField
                label="Teacher"
                required
                value={editForm.teacher_user_id}
                teachers={assignFormTeachers}
                lockedTeacher={lockedAssignTeacher}
                onChange={(e) => setEditForm((p) => ({ ...p, teacher_user_id: e.target.value }))}
              />
              <div className="grid-3">
                <div className="form-group">
                  <label>Class</label>
                  <select
                    required
                    value={editForm.class_id}
                    onChange={(e) => onEditClassChange(e.target.value)}
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
                    required={editHasSections}
                    disabled={!editForm.class_id}
                    value={editForm.section_id}
                    onChange={(e) => onEditSectionChange(e.target.value)}
                  >
                    <option value="">
                      {!editForm.class_id
                        ? "Select class first…"
                        : editHasSections
                          ? "Select section…"
                          : "No sections"}
                    </option>
                    {editSections.map((s) => (
                      <option key={s.id} value={s.id}>{s.name}</option>
                    ))}
                  </select>
                </div>
                <div className="form-group">
                  <label>Subject</label>
                  <select
                    required
                    disabled={!canPickEditSubject}
                    value={editForm.subject_id}
                    onChange={(e) => setEditForm((p) => ({ ...p, subject_id: e.target.value }))}
                  >
                    <option value="">
                      {!editForm.class_id
                        ? "Select class first…"
                        : editHasSections && !editForm.section_id
                          ? "Select section first…"
                          : editAvailableSubjects.length === 0
                            ? "No subjects available"
                            : "Select subject…"}
                    </option>
                    {editAvailableSubjects.map((s) => (
                      <option key={s.id} value={s.id}>
                        {s.name}{s.code ? ` (${s.code})` : ""}
                        {!s.section_id ? " — class-wide" : ""}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
              <div className="btn-row">
                <button type="button" className="btn btn-ghost" onClick={closeEdit}>Cancel</button>
                <button type="submit" className="btn btn-primary" disabled={busy || !editForm.subject_id}>
                  {busy ? "Saving…" : "Save changes"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </>
  );
}
