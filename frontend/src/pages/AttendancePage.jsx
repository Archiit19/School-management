import { useState, useEffect, useCallback } from "react";
import { attendanceApi, academicApi, studentApi } from "../api/client";
import { useAuth } from "../context/AuthContext";

const STATUS_OPTIONS = ["present", "absent", "late", "excused"];

export default function AttendancePage() {
  const { user, hasPerm } = useAuth();
  const isTeacher = user?.role_name === "teacher";
  const canStudentAttendance = hasPerm("mark_attendance") || hasPerm("view_attendance");
  const canTeacherAttendance =
    hasPerm("mark_teacher_attendance") ||
    hasPerm("view_teacher_attendance") ||
    hasPerm("mark_own_teacher_attendance");

  const [tab, setTab] = useState(canStudentAttendance ? "mark-student" : "mark-teacher");
  const [classes, setClasses] = useState([]);
  const [teacherAssignments, setTeacherAssignments] = useState([]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    if (canStudentAttendance) {
      academicApi.getClasses().then(setClasses).catch(() => {});
      if (isTeacher && user?.id) {
        academicApi
          .getTeacherAssignments({ teacher_user_id: user.id })
          .then(setTeacherAssignments)
          .catch(() => {});
      }
    }
  }, [canStudentAttendance, isTeacher, user?.id]);

  return (
    <>
      <div className="page-header">
        <h1>Attendance</h1>
        <p>Manage student and teacher attendance records.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="tabs">
        {canStudentAttendance && (
          <>
            <button className={`tab ${tab === "mark-student" ? "active" : ""}`} onClick={() => setTab("mark-student")}>
              Mark Student
            </button>
            <button className={`tab ${tab === "history-student" ? "active" : ""}`} onClick={() => setTab("history-student")}>
              Student History
            </button>
            <button className={`tab ${tab === "stats-student" ? "active" : ""}`} onClick={() => setTab("stats-student")}>
              Student Stats
            </button>
          </>
        )}
        {canTeacherAttendance && (
          <>
            <button className={`tab ${tab === "mark-teacher" ? "active" : ""}`} onClick={() => setTab("mark-teacher")}>
              Mark Teacher
            </button>
            <button className={`tab ${tab === "history-teacher" ? "active" : ""}`} onClick={() => setTab("history-teacher")}>
              Teacher History
            </button>
            <button className={`tab ${tab === "stats-teacher" ? "active" : ""}`} onClick={() => setTab("stats-teacher")}>
              Teacher Stats
            </button>
          </>
        )}
      </div>

      {tab === "mark-student" && canStudentAttendance && (
        <MarkStudentAttendanceTab
          classes={classes}
          teacherAssignments={teacherAssignments}
          isTeacher={isTeacher}
          setError={setError}
          setSuccess={setSuccess}
        />
      )}
      {tab === "history-student" && canStudentAttendance && (
        <StudentHistoryTab classes={classes} setError={setError} />
      )}
      {tab === "stats-student" && canStudentAttendance && (
        <StudentStatsTab classes={classes} setError={setError} />
      )}
      {tab === "mark-teacher" && canTeacherAttendance && (
        <MarkTeacherAttendanceTab setError={setError} setSuccess={setSuccess} />
      )}
      {tab === "history-teacher" && canTeacherAttendance && (
        <TeacherHistoryTab setError={setError} />
      )}
      {tab === "stats-teacher" && canTeacherAttendance && (
        <TeacherStatsTab setError={setError} />
      )}
    </>
  );
}

function MarkStudentAttendanceTab({ classes, teacherAssignments, isTeacher, setError, setSuccess }) {
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
  const availableSections = selectedClass ? allSections.filter((s) => s.class_id === selectedClass) : [];

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
      list.forEach((s) => {
        initial[s.id] = "present";
      });
      setStatuses(initial);
      setLoaded(true);
    } catch (err) {
      setError(err.message);
    }
  }, [selectedClass, selectedSection, setError]);

  async function handleSubmit() {
    if (!date) {
      setError("Please select a date.");
      return;
    }
    if (students.length === 0) {
      setError("No students to mark.");
      return;
    }
    setError("");
    setBusy(true);
    try {
      const entries = students.map((s) => ({
        student_id: s.id,
        status: statuses[s.id] || "present",
      }));
      const payload = { class_id: selectedClass, date, entries };
      if (selectedSection) payload.section_id = selectedSection;
      const res = await attendanceApi.bulkCreate(payload);
      setSuccess(`Attendance saved: ${res.created} marked, ${res.skipped} skipped.`);
      setTimeout(() => setSuccess(""), 5000);
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <div className="card">
        <div className="card-title">Step 1: Select Class & Date</div>
        <div className="grid-3">
          <div className="form-group">
            <label>Class {isTeacher && <span className="text-muted">(your assigned classes)</span>}</label>
            <select
              value={selectedClass}
              onChange={(e) => {
                setSelectedClass(e.target.value);
                setSelectedSection("");
                setLoaded(false);
                setStudents([]);
              }}
            >
              <option value="">Select class...</option>
              {availableClasses.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>
          <div className="form-group">
            <label>Section (optional)</label>
            <select
              value={selectedSection}
              onChange={(e) => {
                setSelectedSection(e.target.value);
                setLoaded(false);
                setStudents([]);
              }}
            >
              <option value="">All sections</option>
              {availableSections.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
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
          <div className="card-title">Step 2: Mark Attendance ({students.length} students)</div>
          {students.length === 0 ? (
            <div className="empty">No students found for this class/section.</div>
          ) : (
            <>
              <div style={{ display: "flex", gap: 12, alignItems: "center", marginBottom: 12, flexWrap: "wrap" }}>
                {STATUS_OPTIONS.map((status) => (
                  <button
                    key={status}
                    className="btn btn-ghost btn-sm"
                    onClick={() => {
                      const updated = {};
                      students.forEach((s) => {
                        updated[s.id] = status;
                      });
                      setStatuses(updated);
                    }}
                  >
                    Mark All {status}
                  </button>
                ))}
              </div>

              <div className="table-wrap">
                <table>
                  <thead>
                    <tr>
                      <th>#</th>
                      <th>Student Name</th>
                      <th>Student ID</th>
                      <th>Status</th>
                    </tr>
                  </thead>
                  <tbody>
                    {students.map((s, i) => (
                      <tr key={s.id}>
                        <td>{i + 1}</td>
                        <td>
                          <strong>
                            {s.first_name} {s.last_name}
                          </strong>
                        </td>
                        <td>
                          <span className="mono truncate">{s.id}</span>
                        </td>
                        <td>
                          <select
                            value={statuses[s.id] || "present"}
                            onChange={(e) => setStatuses((prev) => ({ ...prev, [s.id]: e.target.value }))}
                          >
                            {STATUS_OPTIONS.map((st) => (
                              <option key={st} value={st}>
                                {st}
                              </option>
                            ))}
                          </select>
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

function StudentHistoryTab({ classes, setError }) {
  const [records, setRecords] = useState([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({ page: 1, limit: 20, date: "", student_id: "", class_id: "", status: "" });
  const [editId, setEditId] = useState(null);
  const [editForm, setEditForm] = useState({ status: "present", remarks: "" });
  const [busy, setBusy] = useState(false);
  const flatClasses = classes.map((c) => c.class || c);

  const load = useCallback(async () => {
    try {
      const res = await attendanceApi.list(query);
      setRecords(res.attendance || []);
      setTotal(res.total || 0);
    } catch (err) {
      setError(err.message);
    }
  }, [query, setError]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleUpdate(e) {
    e.preventDefault();
    setBusy(true);
    try {
      await attendanceApi.update(editId, editForm);
      setEditId(null);
      load();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      {editId && (
        <div className="card">
          <div className="card-title">Edit Student Attendance</div>
          <form onSubmit={handleUpdate}>
            <div className="grid-3">
              <div className="form-group">
                <label>ID</label>
                <input readOnly value={editId} className="mono" />
              </div>
              <div className="form-group">
                <label>Status</label>
                <select value={editForm.status} onChange={(e) => setEditForm((p) => ({ ...p, status: e.target.value }))}>
                  {STATUS_OPTIONS.map((st) => (
                    <option key={st} value={st}>
                      {st}
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Remarks</label>
                <input value={editForm.remarks} onChange={(e) => setEditForm((p) => ({ ...p, remarks: e.target.value }))} />
              </div>
            </div>
            <div className="btn-row">
              <button className="btn btn-primary" disabled={busy}>
                Update
              </button>
              <button type="button" className="btn btn-ghost" onClick={() => setEditId(null)}>
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="card">
        <div className="card-title">Student Attendance Records ({total})</div>
        <div className="grid-4 mb-4">
          <div className="form-group">
            <label>Date</label>
            <input type="date" value={query.date} onChange={(e) => setQuery((q) => ({ ...q, date: e.target.value, page: 1 }))} />
          </div>
          <div className="form-group">
            <label>Student ID</label>
            <input
              placeholder="UUID..."
              value={query.student_id}
              onChange={(e) => setQuery((q) => ({ ...q, student_id: e.target.value, page: 1 }))}
            />
          </div>
          <div className="form-group">
            <label>Class</label>
            <select value={query.class_id} onChange={(e) => setQuery((q) => ({ ...q, class_id: e.target.value, page: 1 }))}>
              <option value="">All</option>
              {flatClasses.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>
          <div className="form-group">
            <label>Status</label>
            <select value={query.status} onChange={(e) => setQuery((q) => ({ ...q, status: e.target.value, page: 1 }))}>
              <option value="">All</option>
              {STATUS_OPTIONS.map((st) => (
                <option key={st} value={st}>
                  {st}
                </option>
              ))}
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Student ID</th>
                <th>Status</th>
                <th>Remarks</th>
                <th>Teacher ID</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {records.length === 0 && (
                <tr>
                  <td colSpan={6} className="empty">
                    No records.
                  </td>
                </tr>
              )}
              {records.map((r) => (
                <tr key={r.id}>
                  <td>{r.date?.split("T")[0]}</td>
                  <td>
                    <span className="mono truncate">{r.student_id}</span>
                  </td>
                  <td>
                    <span className={`status status-${r.status}`}>{r.status}</span>
                  </td>
                  <td>{r.remarks || "-"}</td>
                  <td>
                    <span className="mono truncate">{r.teacher_user_id}</span>
                  </td>
                  <td>
                    <button
                      className="btn btn-ghost btn-sm"
                      onClick={() => {
                        setEditId(r.id);
                        setEditForm({ status: r.status, remarks: r.remarks || "" });
                      }}
                    >
                      Edit
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

function MarkTeacherAttendanceTab({ setError, setSuccess }) {
  const { user, hasPerm } = useAuth();
  const canMarkAny = hasPerm("mark_teacher_attendance");
  const [form, setForm] = useState({
    teacher_user_id: "",
    date: new Date().toISOString().split("T")[0],
    status: "present",
    remarks: "",
  });
  const [busy, setBusy] = useState(false);

  async function submit(e) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      const payload = {
        date: form.date,
        status: form.status,
        remarks: form.remarks,
      };
      if (canMarkAny && form.teacher_user_id) payload.teacher_user_id = form.teacher_user_id;
      await attendanceApi.createTeacher(payload);
      setSuccess("Teacher attendance saved.");
      setTimeout(() => setSuccess(""), 4000);
      setForm((p) => ({ ...p, remarks: "" }));
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="card">
      <div className="card-title">Mark Teacher Attendance</div>
      <form onSubmit={submit}>
        <div className="grid-4">
          {canMarkAny ? (
            <div className="form-group">
              <label>Teacher User ID (optional for self)</label>
              <input
                value={form.teacher_user_id}
                onChange={(e) => setForm((p) => ({ ...p, teacher_user_id: e.target.value }))}
                placeholder={user?.id || "teacher UUID"}
              />
            </div>
          ) : (
            <div className="form-group">
              <label>Teacher User ID</label>
              <input readOnly value={user?.id || ""} className="mono" />
            </div>
          )}
          <div className="form-group">
            <label>Date</label>
            <input type="date" value={form.date} onChange={(e) => setForm((p) => ({ ...p, date: e.target.value }))} />
          </div>
          <div className="form-group">
            <label>Status</label>
            <select value={form.status} onChange={(e) => setForm((p) => ({ ...p, status: e.target.value }))}>
              {STATUS_OPTIONS.map((st) => (
                <option key={st} value={st}>
                  {st}
                </option>
              ))}
            </select>
          </div>
          <div className="form-group">
            <label>Remarks</label>
            <input value={form.remarks} onChange={(e) => setForm((p) => ({ ...p, remarks: e.target.value }))} />
          </div>
        </div>
        <div className="btn-row">
          <button className="btn btn-primary" disabled={busy}>
            {busy ? "Saving..." : "Save Teacher Attendance"}
          </button>
        </div>
      </form>
    </div>
  );
}

function TeacherHistoryTab({ setError }) {
  const { user, hasPerm } = useAuth();
  const canViewAll = hasPerm("view_teacher_attendance") || hasPerm("mark_teacher_attendance");
  const [records, setRecords] = useState([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({
    page: 1,
    limit: 20,
    date: "",
    teacher_user_id: canViewAll ? "" : user?.id || "",
    status: "",
  });
  const [editId, setEditId] = useState(null);
  const [editForm, setEditForm] = useState({ status: "present", remarks: "" });

  const load = useCallback(async () => {
    try {
      const res = await attendanceApi.listTeacher(query);
      setRecords(res.attendance || []);
      setTotal(res.total || 0);
    } catch (err) {
      setError(err.message);
    }
  }, [query, setError]);

  useEffect(() => {
    load();
  }, [load]);

  async function handleUpdate(e) {
    e.preventDefault();
    try {
      await attendanceApi.updateTeacher(editId, editForm);
      setEditId(null);
      load();
    } catch (err) {
      setError(err.message);
    }
  }

  return (
    <>
      {editId && (
        <div className="card">
          <div className="card-title">Edit Teacher Attendance</div>
          <form onSubmit={handleUpdate}>
            <div className="grid-3">
              <div className="form-group">
                <label>ID</label>
                <input readOnly value={editId} className="mono" />
              </div>
              <div className="form-group">
                <label>Status</label>
                <select value={editForm.status} onChange={(e) => setEditForm((p) => ({ ...p, status: e.target.value }))}>
                  {STATUS_OPTIONS.map((st) => (
                    <option key={st} value={st}>
                      {st}
                    </option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Remarks</label>
                <input value={editForm.remarks} onChange={(e) => setEditForm((p) => ({ ...p, remarks: e.target.value }))} />
              </div>
            </div>
            <div className="btn-row">
              <button className="btn btn-primary">Update</button>
              <button type="button" className="btn btn-ghost" onClick={() => setEditId(null)}>
                Cancel
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="card">
        <div className="card-title">Teacher Attendance Records ({total})</div>
        <div className="grid-3 mb-4">
          <div className="form-group">
            <label>Date</label>
            <input type="date" value={query.date} onChange={(e) => setQuery((q) => ({ ...q, date: e.target.value, page: 1 }))} />
          </div>
          <div className="form-group">
            <label>Teacher User ID</label>
            <input
              disabled={!canViewAll}
              value={query.teacher_user_id}
              onChange={(e) => setQuery((q) => ({ ...q, teacher_user_id: e.target.value, page: 1 }))}
              placeholder="UUID"
            />
          </div>
          <div className="form-group">
            <label>Status</label>
            <select value={query.status} onChange={(e) => setQuery((q) => ({ ...q, status: e.target.value, page: 1 }))}>
              <option value="">All</option>
              {STATUS_OPTIONS.map((st) => (
                <option key={st} value={st}>
                  {st}
                </option>
              ))}
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Teacher User ID</th>
                <th>Status</th>
                <th>Remarks</th>
                <th>Recorded By</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {records.length === 0 && (
                <tr>
                  <td colSpan={6} className="empty">
                    No records.
                  </td>
                </tr>
              )}
              {records.map((r) => (
                <tr key={r.id}>
                  <td>{r.date?.split("T")[0]}</td>
                  <td>
                    <span className="mono truncate">{r.teacher_user_id}</span>
                  </td>
                  <td>
                    <span className={`status status-${r.status}`}>{r.status}</span>
                  </td>
                  <td>{r.remarks || "-"}</td>
                  <td>
                    <span className="mono truncate">{r.recorded_by_user_id}</span>
                  </td>
                  <td>
                    <button
                      className="btn btn-ghost btn-sm"
                      onClick={() => {
                        setEditId(r.id);
                        setEditForm({ status: r.status, remarks: r.remarks || "" });
                      }}
                    >
                      Edit
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

function StudentStatsTab({ classes, setError }) {
  const today = new Date();
  const firstOfMonth = new Date(today.getFullYear(), today.getMonth(), 1).toISOString().split("T")[0];
  const [query, setQuery] = useState({
    class_id: "",
    student_id: "",
    start_date: firstOfMonth,
    end_date: today.toISOString().split("T")[0],
  });
  const [stats, setStats] = useState([]);
  const [dateRange, setDateRange] = useState({ start: "", end: "" });
  const [loading, setLoading] = useState(false);
  const flatClasses = classes.map((c) => c.class || c);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await attendanceApi.stats(query);
      setStats(res.stats || []);
      setDateRange({ start: res.start_date, end: res.end_date });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [query, setError]);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div className="card">
      <div className="card-title">Student Attendance Statistics</div>
      <div className="grid-4 mb-4">
        <div className="form-group">
          <label>Class</label>
          <select value={query.class_id} onChange={(e) => setQuery((q) => ({ ...q, class_id: e.target.value }))}>
            <option value="">All Classes</option>
            {flatClasses.map((c) => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
        </div>
        <div className="form-group">
          <label>Student ID</label>
          <input
            placeholder="UUID (optional)"
            value={query.student_id}
            onChange={(e) => setQuery((q) => ({ ...q, student_id: e.target.value }))}
          />
        </div>
        <div className="form-group">
          <label>Start Date</label>
          <input type="date" value={query.start_date} onChange={(e) => setQuery((q) => ({ ...q, start_date: e.target.value }))} />
        </div>
        <div className="form-group">
          <label>End Date</label>
          <input type="date" value={query.end_date} onChange={(e) => setQuery((q) => ({ ...q, end_date: e.target.value }))} />
        </div>
      </div>
      <div className="btn-row mb-4">
        <button className="btn btn-primary" onClick={load} disabled={loading}>
          {loading ? "Loading..." : "Refresh Stats"}
        </button>
      </div>
      {dateRange.start && (
        <p className="text-muted mb-2">
          Showing stats from <strong>{dateRange.start}</strong> to <strong>{dateRange.end}</strong>
        </p>
      )}
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Student ID</th>
              <th>Total Days</th>
              <th>Present</th>
              <th>Absent</th>
              <th>Late</th>
              <th>Excused</th>
              <th>Attendance %</th>
            </tr>
          </thead>
          <tbody>
            {stats.length === 0 && (
              <tr>
                <td colSpan={7} className="empty">No attendance data found for the selected filters.</td>
              </tr>
            )}
            {stats.map((s, i) => (
              <tr key={s.student_id || i}>
                <td><span className="mono truncate">{s.student_id}</span></td>
                <td>{s.total_days}</td>
                <td><span className="status status-present">{s.present_days}</span></td>
                <td><span className="status status-absent">{s.absent_days}</span></td>
                <td><span className="status status-late">{s.late_days}</span></td>
                <td><span className="status status-excused">{s.excused_days}</span></td>
                <td>
                  <strong style={{ color: s.attendance_rate >= 75 ? "var(--success)" : s.attendance_rate >= 50 ? "var(--warning)" : "var(--error)" }}>
                    {s.attendance_rate.toFixed(1)}%
                  </strong>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function TeacherStatsTab({ setError }) {
  const { user, hasPerm } = useAuth();
  const canViewAll = hasPerm("view_teacher_attendance") || hasPerm("mark_teacher_attendance");
  const today = new Date();
  const firstOfMonth = new Date(today.getFullYear(), today.getMonth(), 1).toISOString().split("T")[0];
  const [query, setQuery] = useState({
    teacher_user_id: canViewAll ? "" : user?.id || "",
    start_date: firstOfMonth,
    end_date: today.toISOString().split("T")[0],
  });
  const [stats, setStats] = useState([]);
  const [dateRange, setDateRange] = useState({ start: "", end: "" });
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await attendanceApi.statsTeacher(query);
      setStats(res.stats || []);
      setDateRange({ start: res.start_date, end: res.end_date });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [query, setError]);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div className="card">
      <div className="card-title">Teacher Attendance Statistics</div>
      <div className="grid-3 mb-4">
        <div className="form-group">
          <label>Teacher User ID</label>
          <input
            disabled={!canViewAll}
            placeholder="UUID (optional)"
            value={query.teacher_user_id}
            onChange={(e) => setQuery((q) => ({ ...q, teacher_user_id: e.target.value }))}
          />
        </div>
        <div className="form-group">
          <label>Start Date</label>
          <input type="date" value={query.start_date} onChange={(e) => setQuery((q) => ({ ...q, start_date: e.target.value }))} />
        </div>
        <div className="form-group">
          <label>End Date</label>
          <input type="date" value={query.end_date} onChange={(e) => setQuery((q) => ({ ...q, end_date: e.target.value }))} />
        </div>
      </div>
      <div className="btn-row mb-4">
        <button className="btn btn-primary" onClick={load} disabled={loading}>
          {loading ? "Loading..." : "Refresh Stats"}
        </button>
      </div>
      {dateRange.start && (
        <p className="text-muted mb-2">
          Showing stats from <strong>{dateRange.start}</strong> to <strong>{dateRange.end}</strong>
        </p>
      )}
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Teacher User ID</th>
              <th>Total Days</th>
              <th>Present</th>
              <th>Absent</th>
              <th>Late</th>
              <th>Excused</th>
              <th>Attendance %</th>
            </tr>
          </thead>
          <tbody>
            {stats.length === 0 && (
              <tr>
                <td colSpan={7} className="empty">No attendance data found for the selected filters.</td>
              </tr>
            )}
            {stats.map((s, i) => (
              <tr key={s.teacher_user_id || i}>
                <td><span className="mono truncate">{s.teacher_user_id}</span></td>
                <td>{s.total_days}</td>
                <td><span className="status status-present">{s.present_days}</span></td>
                <td><span className="status status-absent">{s.absent_days}</span></td>
                <td><span className="status status-late">{s.late_days}</span></td>
                <td><span className="status status-excused">{s.excused_days}</span></td>
                <td>
                  <strong style={{ color: s.attendance_rate >= 75 ? "var(--success)" : s.attendance_rate >= 50 ? "var(--warning)" : "var(--error)" }}>
                    {s.attendance_rate.toFixed(1)}%
                  </strong>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
