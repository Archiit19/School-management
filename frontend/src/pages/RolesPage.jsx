import { useState, useEffect, useCallback } from "react";
import { rolesApi, permissionsApi } from "../api/client";

export default function RolesPage() {
  const [tab, setTab] = useState("roles");
  const [roles, setRoles] = useState([]);
  const [permissions, setPermissions] = useState([]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const [roleForm, setRoleForm] = useState({ name: "", description: "" });
  const [assignForm, setAssignForm] = useState({ role_id: "", permission_id: "" });
  const [rolePerms, setRolePerms] = useState({});

  const loadRoles = useCallback(async () => {
    try { setRoles(await rolesApi.list()); } catch (err) { setError(err.message); }
  }, []);

  const loadPerms = useCallback(async () => {
    try { setPermissions(await permissionsApi.list()); } catch (err) { setError(err.message); }
  }, []);

  useEffect(() => { loadRoles(); loadPerms(); }, [loadRoles, loadPerms]);

  function msg(txt) { setSuccess(txt); setError(""); setTimeout(() => setSuccess(""), 3000); }

  async function createRole(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try { await rolesApi.create(roleForm); msg("Role created."); setRoleForm({ name: "", description: "" }); loadRoles(); }
    catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  async function assignPerm(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try { await permissionsApi.assign(assignForm); msg("Permission assigned to role."); viewRolePerms(assignForm.role_id); }
    catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  async function viewRolePerms(roleId) {
    try {
      const perms = await permissionsApi.forRole(roleId);
      setRolePerms((p) => ({ ...p, [roleId]: perms }));
    } catch {}
  }

  const PERM_GROUPS = groupPermissions(permissions);

  return (
    <>
      <div className="page-header">
        <h1>Roles & Permissions</h1>
        <p>Flow 2 — Create roles, view predefined permissions, assign permissions to roles.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="tabs">
        <button className={`tab ${tab === "roles" ? "active" : ""}`} onClick={() => setTab("roles")}>Roles</button>
        <button className={`tab ${tab === "permissions" ? "active" : ""}`} onClick={() => setTab("permissions")}>Permissions</button>
        <button className={`tab ${tab === "assign" ? "active" : ""}`} onClick={() => setTab("assign")}>Assign to Role</button>
      </div>

      {tab === "roles" && (
        <>
          <div className="card">
            <div className="card-title">Create Role <span className="badge badge-post">POST</span></div>
            <form onSubmit={createRole}>
              <div className="grid-2">
                <div className="form-group"><label>Name</label><input name="name" required value={roleForm.name} onChange={(e) => setRoleForm((p) => ({ ...p, name: e.target.value }))} placeholder="teacher" /></div>
                <div className="form-group"><label>Description</label><input name="description" value={roleForm.description} onChange={(e) => setRoleForm((p) => ({ ...p, description: e.target.value }))} placeholder="Teacher role" /></div>
              </div>
              <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Create Role</button></div>
            </form>
          </div>
          <div className="card">
            <div className="card-title">Roles <span className="badge badge-get">GET</span></div>
            <div className="table-wrap">
              <table>
                <thead><tr><th>Name</th><th>Description</th><th>ID</th><th>Permissions</th></tr></thead>
                <tbody>
                  {roles.length === 0 && <tr><td colSpan={4} className="empty">No roles.</td></tr>}
                  {roles.map((r) => (
                    <tr key={r.id}>
                      <td><strong>{r.name}</strong></td>
                      <td>{r.description}</td>
                      <td><span className="mono truncate">{r.id}</span></td>
                      <td>
                        {!rolePerms[r.id] ? (
                          <button className="btn btn-ghost btn-sm" onClick={() => viewRolePerms(r.id)}>Load</button>
                        ) : rolePerms[r.id].length === 0 ? (
                          <em className="text-muted text-sm">None assigned</em>
                        ) : (
                          <div style={{ display: "flex", flexWrap: "wrap", gap: 4 }}>
                            {rolePerms[r.id].map((p) => (
                              <span key={p.id} className="status status-active">{p.name}</span>
                            ))}
                          </div>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {tab === "permissions" && (
        <div className="card">
          <div className="card-title">Predefined Permissions <span className="badge badge-get">GET</span></div>
          <p className="text-sm text-muted mb-4">These permissions are seeded automatically. Assign them to roles under the "Assign to Role" tab.</p>
          {Object.entries(PERM_GROUPS).map(([group, perms]) => (
            <div key={group} style={{ marginBottom: 16 }}>
              <div style={{ fontSize: 12, fontWeight: 700, textTransform: "uppercase", color: "var(--clr-text-secondary)", letterSpacing: "0.05em", marginBottom: 6 }}>{group}</div>
              <div className="table-wrap">
                <table>
                  <thead><tr><th>Permission</th><th>Description</th><th>ID</th></tr></thead>
                  <tbody>
                    {perms.map((p) => (
                      <tr key={p.id}>
                        <td><code style={{ background: "#f1f5f9", padding: "2px 6px", borderRadius: 4 }}>{p.name}</code></td>
                        <td>{p.description}</td>
                        <td><span className="mono truncate">{p.id}</span></td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ))}
          {permissions.length === 0 && <div className="empty">No permissions found. Make sure user-service is running.</div>}
        </div>
      )}

      {tab === "assign" && (
        <div className="card">
          <div className="card-title">Assign Permission to Role <span className="badge badge-post">POST</span></div>
          <form onSubmit={assignPerm}>
            <div className="grid-2">
              <div className="form-group">
                <label>Role</label>
                <select required value={assignForm.role_id} onChange={(e) => setAssignForm((p) => ({ ...p, role_id: e.target.value }))}>
                  <option value="">Select role...</option>
                  {roles.map((r) => <option key={r.id} value={r.id}>{r.name}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Permission</label>
                <select required value={assignForm.permission_id} onChange={(e) => setAssignForm((p) => ({ ...p, permission_id: e.target.value }))}>
                  <option value="">Select permission...</option>
                  {permissions.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
                </select>
              </div>
            </div>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Assign</button></div>
          </form>
        </div>
      )}
    </>
  );
}

function groupPermissions(permissions) {
  const groups = {};
  for (const p of permissions) {
    const prefix = p.name.split("_").slice(-1)[0];
    let group;
    if (p.name.match(/user/)) group = "User Management";
    else if (p.name.match(/role|permission/)) group = "Roles & Permissions";
    else if (p.name.match(/class|section|subject|academic/)) group = "Academic Structure";
    else if (p.name.match(/student|admit/)) group = "Students";
    else if (p.name.match(/teacher|assign_teacher/)) group = "Teacher Assignment";
    else if (p.name.match(/attendance/)) group = "Attendance";
    else if (p.name.match(/assignment|submit/)) group = "Assignments";
    else if (p.name.match(/exam|mark|result|publish/)) group = "Exams & Results";
    else if (p.name.match(/fee|payment|due/)) group = "Finance";
    else group = "Other";
    if (!groups[group]) groups[group] = [];
    groups[group].push(p);
  }
  return groups;
}
