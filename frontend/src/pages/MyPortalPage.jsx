import { useState, useEffect, useCallback, createContext, useContext } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import {
  studentApi,
  attendanceApi,
  examApi,
  academicApi,
  financeApi,
  rolesApi,
} from "../api/client";
import PermTabBar from "../components/PermTabBar";
import { usePermTabs } from "../hooks/usePermTabs";

function isAccessDenied(err) {
  return err?.status === 403 || /permission|forbidden/i.test(err?.message || "");
}

const TABS = [
  { id: "profile", label: "Profile", perm: "view_own_profile" },
  { id: "exams", label: "Exams", perm: "view_own_exams" },
  { id: "attendance", label: "Attendance", perm: "view_own_attendance" },
  { id: "results", label: "Results", perm: "view_own_results" },
  { id: "assignments", label: "Assignments", perm: "view_own_assignments" },
  { id: "dues", label: "Dues", perm: "view_own_dues" },
];

function fmtDate(s) {
  if (!s) return "—";
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  return d.toLocaleDateString();
}

function fmtMoney(n) {
  if (n === null || n === undefined) return "—";
  return Number(n).toFixed(2);
}

const PortalChildContext = createContext({
  isParent: false,
  children: [],
  selectedChildId: "",
  setSelectedChildId: () => {},
});

function usePortalChild() {
  return useContext(PortalChildContext);
}

function ChildSelector() {
  const { isParent, children, selectedChildId, setSelectedChildId } = usePortalChild();
  if (!isParent) return null;

  if (children.length === 0) {
    return (
      <div className="alert alert-error" style={{ marginBottom: 16 }}>
        No students are linked to your account yet. Ask the school to assign your children when admitting students.
      </div>
    );
  }

  const selected = children.find((c) => c.id === selectedChildId);

  return (
    <div className="card" style={{ marginBottom: 16 }}>
      <div className="card-title">Viewing information for</div>
      <div className="form-group" style={{ maxWidth: 360, marginBottom: 0 }}>
        <select
          value={selectedChildId}
          onChange={(e) => setSelectedChildId(e.target.value)}
        >
          {children.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
              {c.role_data?.student_code ? ` (${c.role_data.student_code})` : ""}
            </option>
          ))}
        </select>
        {selected && (
          <p className="text-sm" style={{ marginTop: 8, color: "var(--clr-muted)" }}>
            Attendance, exams, assignments, and dues below are for {selected.name}.
          </p>
        )}
      </div>
    </div>
  );
}

function formatRoleFieldValue(key, value, fieldDef, academic) {
  if (value == null || String(value).trim() === "") return "—";
  if (key === "class_id") return academic?.class?.name || String(value);
  if (key === "section_id") return academic?.section?.name || String(value);
  if (fieldDef?.type === "uuid") return String(value);
  return String(value);
}

function roleDetailRows(user, fields, academic) {
  const data = user?.role_data || {};
  const keys = fields?.length ? fields.map((f) => f.key) : Object.keys(data);
  const rows = [];
  const seen = new Set();
  for (const key of keys) {
    if (seen.has(key) || !(key in data)) continue;
    seen.add(key);
    const fieldDef = fields?.find((f) => f.key === key);
    rows.push({
      key,
      label: fieldDef?.label || key.replace(/_/g, " "),
      value: formatRoleFieldValue(key, data[key], fieldDef, academic),
    });
  }
  for (const [key, value] of Object.entries(data)) {
    if (seen.has(key)) continue;
    rows.push({
      key,
      label: key.replace(/_/g, " "),
      value: formatRoleFieldValue(key, value, null, academic),
    });
  }
  return rows;
}

export default function MyPortalPage() {
  const { user } = useAuth();
  const { visibleTabs, tab, setTab } = usePermTabs(TABS, "profile");
  const isStudent = user?.role_name === "student";
  const isParent = user?.role_name === "parent";
  const [children, setChildren] = useState([]);
  const [selectedChildId, setSelectedChildId] = useState("");
  const [childrenLoading, setChildrenLoading] = useState(isParent);

  useEffect(() => {
    if (!isParent) return;
    setChildrenLoading(true);
    studentApi
      .getMyChildren()
      .then((res) => {
        const list = res?.children || [];
        setChildren(list);
        if (list.length > 0) setSelectedChildId(list[0].id);
      })
      .catch(() => setChildren([]))
      .finally(() => setChildrenLoading(false));
  }, [isParent]);

  if (visibleTabs.length === 0) {
    return <Navigate to="/" replace />;
  }

  const portalChildValue = {
    isParent,
    children,
    selectedChildId,
    setSelectedChildId,
  };

  return (
    <PortalChildContext.Provider value={portalChildValue}>
      <div className="page-header">
        <h1>My Portal</h1>
        <p>
          Welcome, {user?.name}.{" "}
          {isStudent
            ? "Here's your profile, class details, and school information."
            : isParent
              ? "Select a child to view their school information."
              : "View your account and school information."}
        </p>
      </div>

      {isParent && childrenLoading && <div className="empty">Loading linked students…</div>}
      {!childrenLoading && <ChildSelector />}

      <PermTabBar tabs={visibleTabs} active={tab} onChange={setTab} />

      {tab === "profile" && <ProfileTab />}
      {tab === "exams" && <ExamsTab />}
      {tab === "attendance" && <AttendanceTab />}
      {tab === "results" && <ResultsTab />}
      {tab === "assignments" && <AssignmentsTab />}
      {tab === "dues" && <DuesTab />}
    </PortalChildContext.Provider>
  );
}

function ExamsTab() {
  const { isParent, selectedChildId } = usePortalChild();
  const [exams, setExams] = useState([]);
  const [upcoming, setUpcoming] = useState(true);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    if (isParent && !selectedChildId) return;
    setError(""); setLoading(true);
    try {
      const data = await examApi.getMyExams({ upcoming }, isParent ? selectedChildId : undefined);
      setExams(data || []);
    } catch (e) {
      if (!isAccessDenied(e)) setError(e.message);
    } finally {
      setLoading(false);
    }
  }, [upcoming, isParent, selectedChildId]);

  useEffect(() => { load(); }, [load]);

  return (
    <>
      {error && <div className="alert alert-error">{error}</div>}
      <div className="card">
        <div className="card-title">
          Exam Schedule <span className="badge badge-get">GET /exams/me</span>
        </div>
        <div className="form-group" style={{ maxWidth: 280, marginBottom: 16 }}>
          <label style={{ display: "flex", alignItems: "center", gap: 8 }}>
            <input
              type="checkbox"
              checked={upcoming}
              onChange={(e) => setUpcoming(e.target.checked)}
            />
            Show only upcoming exams
          </label>
        </div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Title</th><th>Date</th><th>Total Marks</th><th>Status</th></tr></thead>
            <tbody>
              {loading && <tr><td colSpan={4} className="empty">Loading...</td></tr>}
              {!loading && exams.length === 0 && (
                <tr><td colSpan={4} className="empty">{upcoming ? "No upcoming exams scheduled." : "No exams found."}</td></tr>
              )}
              {!loading && exams.map((e) => (
                <tr key={e.id}>
                  <td><strong>{e.title}</strong></td>
                  <td>{fmtDate(e.exam_date)}</td>
                  <td>{e.total_marks}</td>
                  <td>
                    <span className={`status ${e.is_published ? "status-active" : "status-inactive"}`}>
                      {e.is_published ? "Results Published" : "Scheduled"}
                    </span>
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

function ProfileTab() {
  const { user } = useAuth();
  const { isParent, selectedChildId } = usePortalChild();
  const [me, setMe] = useState(null);
  const [childProfile, setChildProfile] = useState(null);
  const [roleFields, setRoleFields] = useState([]);
  const [academic, setAcademic] = useState(null);
  const [loading, setLoading] = useState(true);
  const [academicLoading, setAcademicLoading] = useState(true);
  const [error, setError] = useState("");
  const [academicError, setAcademicError] = useState("");
  const [showAcademic, setShowAcademic] = useState(true);

  useEffect(() => {
    setLoading(true);
    setRoleFields([]);

    studentApi
      .getMe()
      .then(async (profile) => {
        setMe(profile);
        if (profile?.role_id) {
          try {
            const res = await rolesApi.getFields(profile.role_id);
            setRoleFields(res?.fields || []);
          } catch {
            setRoleFields([]);
          }
        }
      })
      .catch((e) => {
        if (!isAccessDenied(e)) setError(e.message);
      })
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (isParent && !selectedChildId) {
      setChildProfile(null);
      setAcademic(null);
      setAcademicLoading(false);
      return;
    }

    setAcademicLoading(true);
    setShowAcademic(true);
    setAcademicError("");

    const pupilId = isParent ? selectedChildId : undefined;
    const profilePromise = isParent
      ? studentApi.getChild(selectedChildId)
      : Promise.resolve(null);

    profilePromise
      .then(async (profile) => {
        setChildProfile(profile);
        if (isParent && profile?.role_id) {
          try {
            const res = await rolesApi.getFields(profile.role_id);
            setRoleFields(res?.fields || []);
          } catch {
            setRoleFields([]);
          }
        }
      })
      .catch(() => setChildProfile(null));

    academicApi
      .getMyAcademic(pupilId)
      .then(setAcademic)
      .catch((e) => {
        if (isAccessDenied(e)) {
          setShowAcademic(false);
          return;
        }
        setAcademicError(e.message || "Failed to load academic details");
      })
      .finally(() => setAcademicLoading(false));
  }, [isParent, selectedChildId]);

  if (loading) return <div className="empty">Loading profile...</div>;
  if (error) return <div className="alert alert-error">{error}</div>;
  if (!me) return null;

  const school = user?.school || {};
  const displayProfile = isParent && childProfile ? childProfile : me;
  const displayRole = displayProfile?.role_name || me.role_name;
  const roleRows = roleDetailRows(displayProfile, roleFields, academic).filter(
    (row) =>
      !(displayRole === "student" && (row.key === "class_id" || row.key === "section_id")) &&
      !(isParent && row.key === "children")
  );
  const className =
    academic?.class?.name ||
    (academicLoading && displayProfile.class_id ? "(loading…)" : "—");
  const sectionName =
    academic?.section?.name ||
    (academicLoading && displayProfile.section_id ? "(loading…)" : "—");

  return (
    <>
      {isParent && (
        <div className="card">
          <div className="card-title">Parent Account</div>
          <div className="grid-3">
            <div className="form-group">
              <label>Your Name</label>
              <input readOnly value={me.name || "—"} />
            </div>
            <div className="form-group">
              <label>Your Email</label>
              <input readOnly value={me.email || "—"} />
            </div>
            <div className="form-group">
              <label>Role</label>
              <input readOnly value={me.role_name || "—"} />
            </div>
          </div>
        </div>
      )}

      <div className="card">
        <div className="card-title">
          {isParent ? "Student Profile" : "My Account"}{" "}
          <span className="badge badge-get">{isParent ? "GET /users/me/children/:id" : "GET /users/me"}</span>
        </div>
        <div className="grid-3">
          <div className="form-group">
            <label>Full Name</label>
            <input readOnly value={displayProfile.name || "—"} />
          </div>
          <div className="form-group">
            <label>Email</label>
            <input readOnly value={displayProfile.email || "—"} />
          </div>
          <div className="form-group">
            <label>Role</label>
            <input readOnly value={displayProfile.role_name || "—"} />
          </div>
          <div className="form-group">
            <label>Status</label>
            <input readOnly value={displayProfile.is_active ? "Active" : "Inactive"} />
          </div>
          <div className="form-group">
            <label>Member Since</label>
            <input readOnly value={fmtDate(displayProfile.created_at)} />
          </div>
          <div className="form-group">
            <label>User ID</label>
            <input readOnly className="mono" value={displayProfile.id || "—"} />
          </div>
        </div>
      </div>

      {(displayRole === "student" || roleRows.length > 0) && (
        <div className="card">
          <div className="card-title">
            {displayRole === "student" ? "Student Details" : "Role Information"}
          </div>
          <div className="grid-3">
            {displayRole === "student" && (
              <>
                <div className="form-group">
                  <label>Class</label>
                  <input readOnly value={className} />
                </div>
                <div className="form-group">
                  <label>Section</label>
                  <input readOnly value={sectionName} />
                </div>
              </>
            )}
            {roleRows.map((row) => (
              <div className="form-group" key={row.key}>
                <label style={{ textTransform: "capitalize" }}>{row.label}</label>
                <input
                  readOnly
                  className={row.key === "student_code" ? "mono" : undefined}
                  value={row.value}
                  style={row.key === "student_code" ? { fontWeight: 600 } : undefined}
                />
              </div>
            ))}
            {roleRows.length === 0 && displayRole !== "student" && (
              <div className="empty">No role-specific information on file.</div>
            )}
          </div>
        </div>
      )}

      <div className="card">
        <div className="card-title">
          School <span className="badge badge-get">GET /auth/me</span>
        </div>
        <div className="grid-3">
          <div className="form-group">
            <label>School Name</label>
            <input readOnly value={school.name || "—"} />
          </div>
          <div className="form-group">
            <label>School Email</label>
            <input readOnly value={school.email || "—"} />
          </div>
          <div className="form-group">
            <label>School Phone</label>
            <input readOnly value={school.phone || "—"} />
          </div>
        </div>
      </div>

      {showAcademic && displayRole === "student" && (
      <div className="card">
        <div className="card-title">
          Subjects & Teachers <span className="badge badge-get">GET /academic/me</span>
        </div>
        {academicLoading && <div className="empty">Loading class details…</div>}
        {!academicLoading && academicError && (
          <div className="alert alert-error" style={{ marginBottom: 16 }}>{academicError}</div>
        )}
        {!academicLoading && !academicError && academic && (
          <>
            <div style={{ marginBottom: 12, fontWeight: 600 }}>Subjects</div>
            {(!academic.subjects || academic.subjects.length === 0) && (
              <div className="empty">No subjects configured for this class.</div>
            )}
            {academic.subjects && academic.subjects.length > 0 && (
              <div className="table-wrap" style={{ marginBottom: 16 }}>
                <table>
                  <thead><tr><th>Subject</th><th>Code</th></tr></thead>
                  <tbody>
                    {academic.subjects.map((s) => (
                      <tr key={s.id}>
                        <td>{s.name}</td>
                        <td>{s.code || "—"}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}

            <div style={{ marginBottom: 12, fontWeight: 600 }}>Class Teachers</div>
            {(!academic.teachers || academic.teachers.length === 0) && (
              <div className="empty">No teachers have been assigned to this class yet.</div>
            )}
            {academic.teachers && academic.teachers.length > 0 && (
              <div className="table-wrap">
                <table>
                  <thead><tr><th>Teacher</th><th>Email</th><th>Subject</th></tr></thead>
                  <tbody>
                    {academic.teachers.map((t, i) => (
                      <tr key={`${t.teacher_user_id}-${t.subject_id}-${i}`}>
                        <td>{t.teacher_name || "—"}</td>
                        <td>{t.teacher_email || "—"}</td>
                        <td>{t.subject_name || "—"}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </>
        )}
        {!academicLoading && !academicError && !academic && (
          <div className="empty">No academic profile available.</div>
        )}
      </div>
      )}
    </>
  );
}

function AttendanceTab() {
  const { isParent, selectedChildId } = usePortalChild();
  const today = new Date().toISOString().slice(0, 10);
  const monthStart = today.slice(0, 8) + "01";
  const [range, setRange] = useState({ start_date: monthStart, end_date: today });
  const [list, setList] = useState([]);
  const [stats, setStats] = useState(null);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    if (isParent && !selectedChildId) return;
    setError("");
    const pupilId = isParent ? selectedChildId : undefined;
    try {
      const [att, st] = await Promise.all([
        attendanceApi.getMine({ start_date: range.start_date, end_date: range.end_date, limit: 100 }, pupilId),
        attendanceApi.myStats({ start_date: range.start_date, end_date: range.end_date }, pupilId),
      ]);
      setList(att?.attendance || []);
      setStats(st?.stats?.[0] || { total_days: 0, present_days: 0, absent_days: 0, late_days: 0, excused_days: 0, attendance_rate: 0 });
    } catch (e) {
      if (!isAccessDenied(e)) setError(e.message);
    }
  }, [range.start_date, range.end_date, isParent, selectedChildId]);

  useEffect(() => { load(); }, [load]);

  return (
    <>
      <div className="card">
        <div className="card-title">
          Date Range <span className="badge badge-get">GET /attendance/me</span>
        </div>
        <div className="grid-3">
          <div className="form-group">
            <label>Start</label>
            <input type="date" value={range.start_date} onChange={(e) => setRange((r) => ({ ...r, start_date: e.target.value }))} />
          </div>
          <div className="form-group">
            <label>End</label>
            <input type="date" value={range.end_date} onChange={(e) => setRange((r) => ({ ...r, end_date: e.target.value }))} />
          </div>
        </div>
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      {stats && (
        <div className="card">
          <div className="card-title">Summary</div>
          <div className="stats-row">
            <div className="stat-card"><div className="label">Total Days</div><div className="value">{stats.total_days}</div></div>
            <div className="stat-card"><div className="label">Present</div><div className="value" style={{ color: "var(--clr-success)" }}>{stats.present_days}</div></div>
            <div className="stat-card"><div className="label">Absent</div><div className="value" style={{ color: "var(--clr-danger)" }}>{stats.absent_days}</div></div>
            <div className="stat-card"><div className="label">Late</div><div className="value">{stats.late_days}</div></div>
            <div className="stat-card"><div className="label">Excused</div><div className="value">{stats.excused_days}</div></div>
            <div className="stat-card"><div className="label">Rate</div><div className="value">{Number(stats.attendance_rate || 0).toFixed(1)}%</div></div>
          </div>
        </div>
      )}

      <div className="card">
        <div className="card-title">Daily Records</div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Date</th><th>Status</th><th>Remarks</th></tr></thead>
            <tbody>
              {list.length === 0 && <tr><td colSpan={3} className="empty">No records in this range.</td></tr>}
              {list.map((a) => (
                <tr key={a.id}>
                  <td>{fmtDate(a.date)}</td>
                  <td><span className={`status ${a.status === "present" ? "status-active" : a.status === "absent" ? "status-inactive" : ""}`}>{a.status}</span></td>
                  <td>{a.remarks || "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

function ResultsTab() {
  const { isParent, selectedChildId } = usePortalChild();
  const [results, setResults] = useState([]);
  const [error, setError] = useState("");

  useEffect(() => {
    if (isParent && !selectedChildId) return;
    examApi
      .getMyResults({}, isParent ? selectedChildId : undefined)
      .then((r) => setResults(r?.results || r || []))
      .catch((e) => { if (!isAccessDenied(e)) setError(e.message); });
  }, [isParent, selectedChildId]);

  return (
    <div className="card">
      <div className="card-title">
        Published Results <span className="badge badge-get">GET /results/me</span>
      </div>
      {error && <div className="alert alert-error">{error}</div>}
      <div className="table-wrap">
        <table>
          <thead><tr><th>Exam</th><th>Date</th><th>Marks</th><th>Total</th><th>%</th><th>Grade</th></tr></thead>
          <tbody>
            {results.length === 0 && <tr><td colSpan={6} className="empty">No published results yet.</td></tr>}
            {results.map((r) => (
              <tr key={r.exam_id}>
                <td>{r.exam_title}</td>
                <td>{fmtDate(r.exam_date)}</td>
                <td>{r.marks_obtained}</td>
                <td>{r.total_marks}</td>
                <td>{Number(r.percentage || 0).toFixed(1)}%</td>
                <td>{r.grade || "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function AssignmentsTab() {
  const { hasPerm, user } = useAuth();
  const { isParent, selectedChildId } = usePortalChild();
  const canSubmit = hasPerm("submit_own_assignment") && user?.role_name === "student";

  const [assignments, setAssignments] = useState([]);
  const [submissions, setSubmissions] = useState([]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);
  const [active, setActive] = useState(null);
  const [draft, setDraft] = useState({ content: "", material_url: "" });

  const load = useCallback(async () => {
    if (isParent && !selectedChildId) return;
    setError("");
    const pupilId = isParent ? selectedChildId : undefined;
    try {
      const [a, s] = await Promise.all([
        academicApi.getMyAssignments(pupilId),
        academicApi.getMySubmissions(pupilId),
      ]);
      setAssignments(a || []);
      setSubmissions(s || []);
    } catch (e) {
      if (!isAccessDenied(e)) setError(e.message);
    }
  }, [isParent, selectedChildId]);

  useEffect(() => { load(); }, [load]);

  const submittedFor = new Map(submissions.map((s) => [s.assignment_id, s]));
  const assignmentTitle = (id) => assignments.find((a) => a.id === id)?.title || id;

  function openSubmit(a) {
    setActive(a);
    setDraft({ content: "", material_url: "" });
  }

  async function submit(e) {
    e.preventDefault(); setBusy(true); setError("");
    try {
      await academicApi.submitMine({ assignment_id: active.id, ...draft });
      setSuccess("Submitted."); setTimeout(() => setSuccess(""), 3000);
      setActive(null);
      load();
    } catch (e) { setError(e.message); } finally { setBusy(false); }
  }

  return (
    <>
      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="card">
        <div className="card-title">
          Assignments for My Class <span className="badge badge-get">GET /assignments/me</span>
        </div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Title</th><th>Due</th><th>Material</th><th>Status</th><th></th></tr></thead>
            <tbody>
              {assignments.length === 0 && <tr><td colSpan={5} className="empty">No assignments.</td></tr>}
              {assignments.map((a) => {
                const sub = submittedFor.get(a.id);
                return (
                  <tr key={a.id}>
                    <td>{a.title}</td>
                    <td>{a.due_date ? fmtDate(a.due_date) : "—"}</td>
                    <td>{a.material_url ? <a href={a.material_url} target="_blank" rel="noreferrer">Open</a> : "—"}</td>
                    <td>
                      {sub ? (
                        sub.reviewed_at ? (
                          <span className="status status-active">
                            Reviewed{sub.marks != null ? ` · ${sub.marks}/20` : ""}
                          </span>
                        ) : (
                          <span className="status status-active">Submitted</span>
                        )
                      ) : (
                        <span className="status status-inactive">Pending</span>
                      )}
                    </td>
                    <td>
                      {canSubmit && !sub && (
                        <button className="btn btn-primary btn-sm" onClick={() => openSubmit(a)}>Submit</button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      {active && (
        <div className="card">
          <div className="card-title">
            Submit: {active.title} <span className="badge badge-post">POST /submissions/me</span>
          </div>
          <form onSubmit={submit}>
            <div className="form-group">
              <label>Notes / Content</label>
              <textarea rows={5} value={draft.content} onChange={(e) => setDraft((d) => ({ ...d, content: e.target.value }))} placeholder="Write your answer..." />
            </div>
            <div className="form-group">
              <label>Material URL (optional)</label>
              <input type="url" value={draft.material_url} onChange={(e) => setDraft((d) => ({ ...d, material_url: e.target.value }))} placeholder="https://..." />
            </div>
            <div className="btn-row">
              <button className="btn btn-primary" disabled={busy}>{busy ? "Submitting..." : "Send Submission"}</button>
              <button type="button" className="btn btn-ghost" onClick={() => setActive(null)}>Cancel</button>
            </div>
          </form>
        </div>
      )}

      <div className="card">
        <div className="card-title">
          My Submissions <span className="badge badge-get">GET /submissions/me</span>
        </div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Assignment</th><th>Submitted</th><th>Marks</th><th>Feedback</th><th>Material</th></tr></thead>
            <tbody>
              {submissions.length === 0 && <tr><td colSpan={5} className="empty">No submissions yet.</td></tr>}
              {submissions.map((s) => (
                <tr key={s.id}>
                  <td><strong>{assignmentTitle(s.assignment_id)}</strong></td>
                  <td>{fmtDate(s.created_at)}</td>
                  <td>
                    {s.reviewed_at && s.marks != null ? (
                      <strong>{s.marks}/20</strong>
                    ) : s.reviewed_at ? (
                      <span className="text-muted">Reviewed</span>
                    ) : (
                      <span className="text-muted">Pending review</span>
                    )}
                  </td>
                  <td>{s.teacher_feedback || (s.reviewed_at ? "—" : <span className="text-muted">Not yet</span>)}</td>
                  <td>{s.material_url ? <a href={s.material_url} target="_blank" rel="noreferrer">Open</a> : "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

function DuesTab() {
  const { isParent, selectedChildId } = usePortalChild();
  const [dues, setDues] = useState([]);
  const [error, setError] = useState("");

  useEffect(() => {
    if (isParent && !selectedChildId) return;
    financeApi
      .getMyDues(isParent ? selectedChildId : undefined)
      .then((d) => setDues(d?.dues || d || []))
      .catch((e) => { if (!isAccessDenied(e)) setError(e.message); });
  }, [isParent, selectedChildId]);

  const totalBalance = dues.reduce((sum, d) => sum + Number(d.balance || 0), 0);

  return (
    <>
      {error && <div className="alert alert-error">{error}</div>}
      <div className="card">
        <div className="card-title">
          Outstanding Dues <span className="badge badge-get">GET /dues/me</span>
        </div>
        <div className="stats-row">
          <div className="stat-card">
            <div className="label">Total Balance</div>
            <div className="value" style={{ color: totalBalance > 0 ? "var(--clr-danger)" : "var(--clr-success)" }}>{fmtMoney(totalBalance)}</div>
          </div>
          <div className="stat-card"><div className="label">Items</div><div className="value">{dues.length}</div></div>
        </div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Fee</th><th>Amount</th><th>Paid</th><th>Balance</th><th>Status</th><th>Due Date</th></tr></thead>
            <tbody>
              {dues.length === 0 && <tr><td colSpan={6} className="empty">No dues found.</td></tr>}
              {dues.map((d) => (
                <tr key={d.fee_id}>
                  <td>{d.title}</td>
                  <td>{fmtMoney(d.amount)}</td>
                  <td>{fmtMoney(d.paid_amount)}</td>
                  <td>{fmtMoney(d.balance)}</td>
                  <td><span className={`status ${d.status === "paid" ? "status-active" : "status-inactive"}`}>{d.status}</span></td>
                  <td>{d.due_date ? fmtDate(d.due_date) : "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
