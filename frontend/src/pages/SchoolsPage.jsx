import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { schoolApi, authApi } from "../api/client";
import { useAuth } from "../context/AuthContext";

const EMPTY_FORM = { name: "", email: "", address: "", phone: "" };

export default function SchoolsPage() {
  const { saveToken, inSchoolContext } = useAuth();
  const navigate = useNavigate();
  const [schools, setSchools] = useState([]);
  const [form, setForm] = useState(EMPTY_FORM);
  const [showCreate, setShowCreate] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(true);

  function load() {
    setLoading(true);
    schoolApi
      .listMine()
      .then((res) => setSchools(res.schools || []))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }

  useEffect(() => { load(); }, []);

  function msg(txt) {
    setSuccess(txt);
    setError("");
    setTimeout(() => setSuccess(""), 3000);
  }

  async function handleCreate(e) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await schoolApi.create(form);
      msg("School created. Open it to manage users, students, and academics.");
      setForm(EMPTY_FORM);
      setShowCreate(false);
      load();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  async function openSchool(schoolId) {
    setError("");
    setBusy(true);
    try {
      const res = await authApi.selectSchool(schoolId);
      saveToken(res.token);
      navigate("/", { replace: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      <div className="page-header">
        <h1>My Schools</h1>
        <p>Schools you administer. Create a school, then open it to manage day-to-day operations.</p>
      </div>

      {inSchoolContext && (
        <div className="alert alert-success" style={{ marginBottom: 16 }}>
          You are inside a school workspace. Use &quot;Exit school&quot; in the sidebar to return here.
        </div>
      )}

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="card" style={{ marginBottom: 16 }}>
        <div className="card-title" style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          <span>Your schools <span className="badge badge-get">GET /schools/mine</span></span>
          <button type="button" className="btn btn-primary btn-sm" onClick={() => setShowCreate((v) => !v)}>
            {showCreate ? "Cancel" : "+ Create school"}
          </button>
        </div>

        {showCreate && (
          <form onSubmit={handleCreate} style={{ marginBottom: 20, paddingBottom: 20, borderBottom: "1px solid var(--clr-border)" }}>
            <div className="grid-2">
              <div className="form-group">
                <label>School name</label>
                <input required value={form.name} onChange={(e) => setForm((p) => ({ ...p, name: e.target.value }))} placeholder="Springfield Elementary" />
              </div>
              <div className="form-group">
                <label>School email</label>
                <input type="email" required value={form.email} onChange={(e) => setForm((p) => ({ ...p, email: e.target.value }))} placeholder="info@school.edu" />
              </div>
              <div className="form-group">
                <label>Phone</label>
                <input value={form.phone} onChange={(e) => setForm((p) => ({ ...p, phone: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Address</label>
                <input value={form.address} onChange={(e) => setForm((p) => ({ ...p, address: e.target.value }))} />
              </div>
            </div>
            <div className="btn-row">
              <button type="submit" className="btn btn-primary" disabled={busy}>
                {busy ? "Creating…" : "Create school"}
              </button>
            </div>
          </form>
        )}

        {loading && <div className="empty">Loading schools…</div>}
        {!loading && schools.length === 0 && (
          <div className="empty">No schools yet. Click &quot;Create school&quot; to add your first institution.</div>
        )}

        {!loading && schools.length > 0 && (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Phone</th>
                  <th>Status</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {schools.map((s) => (
                  <tr key={s.id}>
                    <td><strong>{s.name}</strong></td>
                    <td>{s.email}</td>
                    <td>{s.phone || "—"}</td>
                    <td>
                      <span className={`status ${s.is_active ? "status-active" : "status-inactive"}`}>
                        {s.is_active ? "Active" : "Inactive"}
                      </span>
                    </td>
                    <td>
                      <button type="button" className="btn btn-primary btn-sm" disabled={busy} onClick={() => openSchool(s.id)}>
                        Open school
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </>
  );
}
