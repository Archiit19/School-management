import { useState, useEffect, useCallback } from "react";
import { useAuth } from "../context/AuthContext";
import {
  studentApi,
  attendanceApi,
  examApi,
  academicApi,
  financeApi,
} from "../api/client";

const TABS = [
  { id: "profile", label: "Profile", perm: "view_own_profile" },
  { id: "attendance", label: "Attendance", perm: "view_own_attendance" },
  { id: "exams", label: "Exams", perm: "view_own_exams" },
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

export default function MyPortalPage() {
  const { user, hasPerm } = useAuth();
  const visibleTabs = TABS.filter((t) => hasPerm(t.perm));
  const [tab, setTab] = useState(visibleTabs[0]?.id || "profile");

  return (
    <>
      <div className="page-header">
        <h1>My Portal</h1>
        <p>
          Welcome, {user?.name}. Here's everything tied to your student record.
        </p>
      </div>

      <div className="tabs">
        {visibleTabs.map((t) => (
          <button
            key={t.id}
            className={`tab ${tab === t.id ? "active" : ""}`}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === "profile" && <ProfileTab />}
      {tab === "attendance" && <AttendanceTab />}
      {tab === "exams" && <ExamsTab />}
      {tab === "results" && <ResultsTab />}
      {tab === "assignments" && <AssignmentsTab />}
      {tab === "dues" && <DuesTab />}
    </>
  );
}

function ProfileTab() {
  const [me, setMe] = useState(null);
  const [classes, setClasses] = useState([]);
  const [error, setError] = useState("");

  useEffect(() => {
    studentApi.getMe().then(setMe).catch((e) => setError(e.message));
    academicApi.getClasses().then(setClasses).catch(() => {});
  }, []);

  const flatClasses = classes.map((c) => c.class || c);
  const flatSections = classes.flatMap((c) => (c.sections || []).map((s) => ({ ...s, className: (c.class || c).name })));
  const classMap = Object.fromEntries(flatClasses.map((c) => [c.id, c.name]));
  const sectionMap = Object.fromEntries(flatSections.map((s) => [s.id, s.name]));

  if (error) return <div className="alert alert-error">{error}</div>;
  if (!me) return <div className="empty">Loading...</div>;

  return (
    <div className="card">
      <div className="card-title">
        Student Record <span className="badge badge-get">GET /students/me</span>
      </div>
      <div className="grid-3">
        <div className="form-group">
          <label>First Name</label>
          <input readOnly value={me.first_name || ""} />
        </div>
        <div className="form-group">
          <label>Last Name</label>
          <input readOnly value={me.last_name || ""} />
        </div>
        <div className="form-group">
          <label>Status</label>
          <input readOnly value={me.is_active ? "Active" : "Inactive"} />
        </div>
        <div className="form-group">
          <label>Student ID</label>
          <input readOnly className="mono" value={me.id || ""} />
        </div>
        <div className="form-group">
          <label>Class</label>
          <input readOnly value={classMap[me.class_id] || me.class_id || ""} />
        </div>
        <div className="form-group">
          <label>Section</label>
          <input readOnly value={me.section_id ? (sectionMap[me.section_id] || me.section_id) : "—"} />
        </div>
        <div className="form-group">
          <label>Admitted</label>
          <input readOnly value={fmtDate(me.created_at)} />
        </div>
      </div>
    </div>
  );
}

function AttendanceTab() {
  const today = new Date().toISOString().slice(0, 10);
  const monthStart = today.slice(0, 8) + "01";
  const [range, setRange] = useState({ start_date: monthStart, end_date: today });
  const [list, setList] = useState([]);
  const [stats, setStats] = useState(null);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setError("");
    try {
      const [att, st] = await Promise.all([
        attendanceApi.getMine({ start_date: range.start_date, end_date: range.end_date, limit: 100 }),
        attendanceApi.myStats({ start_date: range.start_date, end_date: range.end_date }),
      ]);
      setList(att?.attendance || []);
      setStats(st?.stats?.[0] || { total_days: 0, present_days: 0, absent_days: 0, late_days: 0, excused_days: 0, attendance_rate: 0 });
    } catch (e) {
      setError(e.message);
    }
  }, [range.start_date, range.end_date]);

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

function ExamsTab() {
  const [exams, setExams] = useState([]);
  const [error, setError] = useState("");

  useEffect(() => {
    examApi.getMyExams().then((e) => setExams(e || [])).catch((e) => setError(e.message));
  }, []);

  const today = new Date().toISOString().slice(0, 10);
  const upcoming = exams.filter((e) => !e.is_published && e.exam_date?.slice(0, 10) >= today);
  const completed = exams.filter((e) => e.is_published || e.exam_date?.slice(0, 10) < today);

  return (
    <>
      {error && <div className="alert alert-error">{error}</div>}

      <div className="card">
        <div className="card-title">
          Upcoming Exams <span className="badge badge-get">GET /exams/me</span>
        </div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Exam</th><th>Date</th><th>Total Marks</th><th>Status</th></tr></thead>
            <tbody>
              {upcoming.length === 0 && <tr><td colSpan={4} className="empty">No upcoming exams.</td></tr>}
              {upcoming.map((e) => (
                <tr key={e.id}>
                  <td><strong>{e.title}</strong></td>
                  <td>{fmtDate(e.exam_date)}</td>
                  <td>{e.total_marks}</td>
                  <td><span className="status" style={{ background: "#fef3c7", color: "#92400e" }}>Upcoming</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <div className="card">
        <div className="card-title">Completed Exams</div>
        <div className="table-wrap">
          <table>
            <thead><tr><th>Exam</th><th>Date</th><th>Total Marks</th><th>Status</th></tr></thead>
            <tbody>
              {completed.length === 0 && <tr><td colSpan={4} className="empty">No completed exams.</td></tr>}
              {completed.map((e) => (
                <tr key={e.id}>
                  <td><strong>{e.title}</strong></td>
                  <td>{fmtDate(e.exam_date)}</td>
                  <td>{e.total_marks}</td>
                  <td>
                    <span className={`status ${e.is_published ? "status-active" : "status-inactive"}`}>
                      {e.is_published ? "Results Published" : "Awaiting Results"}
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

function ResultsTab() {
  const [results, setResults] = useState([]);
  const [error, setError] = useState("");

  useEffect(() => {
    examApi.getMyResults().then((r) => setResults(r?.results || r || [])).catch((e) => setError(e.message));
  }, []);

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
  const { hasPerm } = useAuth();
  const canSubmit = hasPerm("submit_own_assignment");

  const [assignments, setAssignments] = useState([]);
  const [submissions, setSubmissions] = useState([]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);
  const [active, setActive] = useState(null);
  const [draft, setDraft] = useState({ content: "", material_url: "" });

  const load = useCallback(async () => {
    setError("");
    try {
      const [a, s] = await Promise.all([
        academicApi.getMyAssignments(),
        academicApi.getMySubmissions(),
      ]);
      setAssignments(a || []);
      setSubmissions(s || []);
    } catch (e) {
      setError(e.message);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  const submittedFor = new Map(submissions.map((s) => [s.assignment_id, s]));

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
                    <td>{sub ? <span className="status status-active">Submitted</span> : <span className="status status-inactive">Pending</span>}</td>
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
            <thead><tr><th>Assignment ID</th><th>Submitted</th><th>Material</th></tr></thead>
            <tbody>
              {submissions.length === 0 && <tr><td colSpan={3} className="empty">No submissions yet.</td></tr>}
              {submissions.map((s) => (
                <tr key={s.id}>
                  <td><span className="mono truncate">{s.assignment_id}</span></td>
                  <td>{fmtDate(s.created_at)}</td>
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
  const [dues, setDues] = useState([]);
  const [error, setError] = useState("");

  useEffect(() => {
    financeApi.getMyDues().then((d) => setDues(d?.dues || d || [])).catch((e) => setError(e.message));
  }, []);

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
