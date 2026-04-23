import { useState, useEffect, useCallback } from "react";
import { attendanceApi, academicApi, studentApi } from "../api/client";
import { useAuth } from "../context/AuthContext";

export default function AttendancePage() {
  const { user } = useAuth();
  const isTeacher = user?.role_name === "teacher";

  const [tab, setTab] = useState("mark");
  const [classes, setClasses] = useState([]);
  const [teacherAssignments, setTeacherAssignments] = useState([]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    academicApi.getClasses().then(setClasses).catch(() => {});
    if (isTeacher && user?.id) {
      academicApi.getTeacherAssignments({ teacher_user_id: user.id }).then(setTeacherAssignments).catch(() => {});
    }
  }, [isTeacher, user?.id]);

  return (
    <>
      <div className="page-header">
        <h1>Attendance</h1>
        <p>Flow 6 — Select a class, mark attendance for all students at once.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="tabs">
        <button className={`tab ${tab === "mark" ? "active" : ""}`} onClick={() => setTab("mark")}>Mark Attendance</button>
        <button className={`tab ${tab === "history" ? "active" : ""}`} onClick={() => setTab("history")}>History</button>
      </div>

      {tab === "mark" && (
        <MarkAttendanceTab
          classes={classes}
          teacherAssignments={teacherAssignments}
          isTeacher={isTeacher}
          setError={setError}
          setSuccess={setSuccess}
        />
      )}

      {tab === "history" && (
        <HistoryTab classes={classes} setError={setError} />
      )}
    </>
  );
}

function MarkAttendanceTab({ classes, teacherAssignments, isTeacher, setError, setSuccess }) {
  const [selectedClass, setSelectedClass] = useState("");
  const [selectedSection, setSelectedSection] = useState("");
  const [date, setDate] = useState(() => new Date().toISOString().split("T")[0]);
  const [students, setStudents] = useState([]);
  const [statuses, setStatuses] = useState({});
  const [loaded, setLoaded] = useState(false);
  const [busy, setBusy] = useState(false);

  const allClasses = classes.map((c) => c.class || c);
  const allSections = classes.flatMap((c) => (c.sections || []).map((s) => ({ ...s, className: (c.class || c).name })));

  const availableClasses = isTeacher
    ? allClasses.filter((c) => teacherAssignments.some((ta) => ta.class_id === c.id))
    : allClasses;

  const availableSections = selectedClass
    ? allSections.filter((s) => s.class_id === selectedClass)
    : [];

  const loadStudents = useCallback(async () => {
    if (!selectedClass) return;
    setLoaded(false);
    setError("");
    try {
      const query = { class_id: selectedClass, limit: 200 };
      if (selectedSection) query.section_id = selectedSection;
      const res = await studentApi.list(query);
      const list = res.students || [];
      setStudents(list);
      const initial = {};
      list.forEach((s) => { initial[s.id] = "present"; });
      setStatuses(initial);
      setLoaded(true);
    } catch (err) { setError(err.message); }
  }, [selectedClass, selectedSection, setError]);

  function handleStatusChange(studentId, status) {
    setStatuses((prev) => ({ ...prev, [studentId]: status }));
  }

  function markAll(status) {
    const updated = {};
    students.forEach((s) => { updated[s.id] = status; });
    setStatuses(updated);
  }

  async function handleSubmit() {
    if (!date) { setError("Please select a date."); return; }
    if (students.length === 0) { setError("No students to mark."); return; }
    setError(""); setBusy(true);
    try {
      const entries = students.map((s) => ({
        student_id: s.id,
        status: statuses[s.id] || "present",
      }));
      const payload = {
        class_id: selectedClass,
        date,
        entries,
      };
      if (selectedSection) payload.section_id = selectedSection;

      const res = await attendanceApi.bulkCreate(payload);
      setSuccess(`Attendance saved: ${res.created} marked, ${res.skipped} skipped (already marked).`);
      setTimeout(() => setSuccess(""), 5000);
    } catch (err) { setError(err.message); }
    finally { setBusy(false); }
  }

  const presentCount = Object.values(statuses).filter((s) => s === "present").length;
  const absentCount = Object.values(statuses).filter((s) => s === "absent").length;

  return (
    <>
      <div className="card">
        <div className="card-title">Step 1: Select Class & Date</div>
        <div className="grid-3">
          <div className="form-group">
            <label>Class {isTeacher && <span className="text-muted">(your assigned classes)</span>}</label>
            <select value={selectedClass} onChange={(e) => { setSelectedClass(e.target.value); setSelectedSection(""); setLoaded(false); setStudents([]); }}>
              <option value="">Select class...</option>
              {availableClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </div>
          <div className="form-group">
            <label>Section (optional)</label>
            <select value={selectedSection} onChange={(e) => { setSelectedSection(e.target.value); setLoaded(false); setStudents([]); }}>
              <option value="">All sections</option>
              {availableSections.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
            </select>
          </div>
          <div className="form-group">
            <label>Date</label>
            <input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
          </div>
        </div>
        <div className="btn-row">
          <button className="btn btn-primary" disabled={!selectedClass} onClick={loadStudents}>
            Load Students
          </button>
        </div>
      </div>

      {loaded && (
        <div className="card">
          <div className="card-title">
            Step 2: Mark Attendance ({students.length} students)
          </div>

          {students.length === 0 ? (
            <div className="empty">No students found for this class/section.</div>
          ) : (
            <>
              <div style={{ display: "flex", gap: 12, alignItems: "center", marginBottom: 12 }}>
                <button className="btn btn-ghost btn-sm" onClick={() => markAll("present")}>Mark All Present</button>
                <button className="btn btn-ghost btn-sm" onClick={() => markAll("absent")}>Mark All Absent</button>
                <div style={{ marginLeft: "auto", display: "flex", gap: 16 }}>
                  <span className="text-sm"><strong style={{ color: "var(--clr-success)" }}>{presentCount}</strong> Present</span>
                  <span className="text-sm"><strong style={{ color: "var(--clr-danger)" }}>{absentCount}</strong> Absent</span>
                </div>
              </div>

              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th style={{ width: 40 }}>#</th>
                      <th>Student Name</th>
                      <th>Student ID</th>
                      <th style={{ width: 200, textAlign: "center" }}>Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {students.map((s, i) => (
                      <tr key={s.id}>
                        <td>{i + 1}</td>
                        <td><strong>{s.first_name} {s.last_name}</strong></td>
                        <td><span className="mono truncate">{s.id}</span></td>
                        <td>
                          <div className="radio-group">
                            <label className={`radio-label ${statuses[s.id] === "present" ? "radio-present" : ""}`}>
                              <input
                                type="radio"
                                name={`status-${s.id}`}
                                value="present"
                                checked={statuses[s.id] === "present"}
                                onChange={() => handleStatusChange(s.id, "present")}
                              />
                              Present
                            </label>
                            <label className={`radio-label ${statuses[s.id] === "absent" ? "radio-absent" : ""}`}>
                              <input
                                type="radio"
                                name={`status-${s.id}`}
                                value="absent"
                                checked={statuses[s.id] === "absent"}
                                onChange={() => handleStatusChange(s.id, "absent")}
                              />
                              Absent
                            </label>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <div className="btn-row" style={{ marginTop: 16 }}>
                <button className="btn btn-primary" disabled={busy} onClick={handleSubmit}>
                  {busy ? "Submitting..." : `Submit Attendance (${students.length} students)`}
                </button>
              </div>
            </>
          )}
        </div>
      )}
    </>
  );
}

function HistoryTab({ classes, setError }) {
  const [records, setRecords] = useState([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({ page: 1, limit: 20, date: "", student_id: "", class_id: "", status: "" });
  const [editId, setEditId] = useState(null);
  const [editForm, setEditForm] = useState({ status: "", remarks: "" });
  const [busy, setBusy] = useState(false);

  const flatClasses = classes.map((c) => c.class || c);

  const load = useCallback(async () => {
    try {
      const res = await attendanceApi.list(query);
      setRecords(res.attendance || []);
      setTotal(res.total || 0);
    } catch (err) { setError(err.message); }
  }, [query, setError]);

  useEffect(() => { load(); }, [load]);

  async function handleUpdate(e) {
    e.preventDefault(); setBusy(true);
    try {
      await attendanceApi.update(editId, editForm);
      setEditId(null);
      load();
    } catch (err) { setError(err.message); }
    finally { setBusy(false); }
  }

  return (
    <>
      {editId && (
        <div className="card">
          <div className="card-title">Edit Record <span className="badge badge-patch">PATCH</span></div>
          <form onSubmit={handleUpdate}>
            <div className="grid-3">
              <div className="form-group"><label>ID</label><input readOnly value={editId} className="mono" /></div>
              <div className="form-group">
                <label>Status</label>
                <select value={editForm.status} onChange={(e) => setEditForm((p) => ({ ...p, status: e.target.value }))}>
                  <option value="present">Present</option>
                  <option value="absent">Absent</option>
                  <option value="late">Late</option>
                </select>
              </div>
              <div className="form-group"><label>Remarks</label><input value={editForm.remarks} onChange={(e) => setEditForm((p) => ({ ...p, remarks: e.target.value }))} /></div>
            </div>
            <div className="btn-row">
              <button className="btn btn-primary" disabled={busy}>Update</button>
              <button type="button" className="btn btn-ghost" onClick={() => setEditId(null)}>Cancel</button>
            </div>
          </form>
        </div>
      )}

      <div className="card">
        <div className="card-title">Records ({total}) <span className="badge badge-get">GET /attendance</span></div>
        <div className="grid-4 mb-4">
          <div className="form-group"><label>Date</label><input type="date" value={query.date} onChange={(e) => setQuery((q) => ({ ...q, date: e.target.value, page: 1 }))} /></div>
          <div className="form-group"><label>Student ID</label><input placeholder="UUID..." value={query.student_id} onChange={(e) => setQuery((q) => ({ ...q, student_id: e.target.value, page: 1 }))} /></div>
          <div className="form-group">
            <label>Class</label>
            <select value={query.class_id} onChange={(e) => setQuery((q) => ({ ...q, class_id: e.target.value, page: 1 }))}>
              <option value="">All</option>
              {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </div>
          <div className="form-group">
            <label>Status</label>
            <select value={query.status} onChange={(e) => setQuery((q) => ({ ...q, status: e.target.value, page: 1 }))}>
              <option value="">All</option>
              <option value="present">Present</option>
              <option value="absent">Absent</option>
              <option value="late">Late</option>
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Date</th><th>Student ID</th><th>Status</th><th>Remarks</th><th>Teacher ID</th><th></th></tr></thead>
            <tbody>
              {records.length === 0 && <tr><td colSpan={6} className="empty">No records.</td></tr>}
              {records.map((r) => (
                <tr key={r.id}>
                  <td>{r.date?.split("T")[0]}</td>
                  <td><span className="mono truncate">{r.student_id}</span></td>
                  <td><span className={`status status-${r.status}`}>{r.status}</span></td>
                  <td>{r.remarks || "—"}</td>
                  <td><span className="mono truncate">{r.teacher_user_id}</span></td>
                  <td><button className="btn btn-ghost btn-sm" onClick={() => { setEditId(r.id); setEditForm({ status: r.status, remarks: r.remarks || "" }); }}>Edit</button></td>
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
