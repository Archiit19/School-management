import { useState, useEffect, useCallback } from "react";
import { rolesApi, permissionsApi } from "../api/client";
import PermGate from "../components/PermGate";
import PermTabBar from "../components/PermTabBar";
import { usePermTabs } from "../hooks/usePermTabs";

const ROLES_TABS = [
  { id: "roles", label: "Roles", any: ["create_role", "manage_permissions"] },
  { id: "permissions", label: "Permissions", perm: "manage_permissions" },
  { id: "assign", label: "Assign to Role", perm: "manage_permissions" },
];

export default function RolesPage() {
  const { visibleTabs, tab, setTab } = usePermTabs(ROLES_TABS, "roles");
  const [roles, setRoles] = useState([]);
  const [permissions, setPermissions] = useState([]);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const [roleForm, setRoleForm] = useState({ name: "", description: "", fields: [] });
  const [newField, setNewField] = useState({ key: "", label: "", type: "text", required: false });
  const [assignForm, setAssignForm] = useState({ role_id: "", permission_id: "" });
  const [rolePerms, setRolePerms] = useState({});
  const [editingRoleId, setEditingRoleId] = useState(null);
  const [addPermId, setAddPermId] = useState("");

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
    try { await rolesApi.create(roleForm); msg("Role created."); setRoleForm({ name: "", description: "", fields: [] }); loadRoles(); }
    catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  function addFieldToRole(e) {
    e.preventDefault();
    if (!newField.key || !newField.label) return;
    setRoleForm((p) => ({ ...p, fields: [...(p.fields || []), { ...newField }] }));
    setNewField({ key: "", label: "", type: "text", required: false });
  }

  function removeField(idx) {
    setRoleForm((p) => ({ ...p, fields: p.fields.filter((_, i) => i !== idx) }));
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
      return perms;
    } catch (err) {
      setError(err.message);
      return [];
    }
  }

  async function openEditPerms(role) {
    setEditingRoleId(role.id);
    setAddPermId("");
    await viewRolePerms(role.id);
  }

  function closeEditPerms() {
    setEditingRoleId(null);
    setAddPermId("");
  }

  async function removePerm(roleId, permissionId) {
    if (!confirm("Remove this permission from the role?")) return;
    setError("");
    setBusy(true);
    try {
      await permissionsApi.removeFromRole(roleId, permissionId);
      msg("Permission removed.");
      await viewRolePerms(roleId);
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  async function addPermToRole(e) {
    e.preventDefault();
    if (!addPermId || !editingRoleId) return;
    setError("");
    setBusy(true);
    try {
      await permissionsApi.assign({ role_id: editingRoleId, permission_id: addPermId });
      msg("Permission added.");
      setAddPermId("");
      await viewRolePerms(editingRoleId);
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  const PERM_GROUPS = groupPermissions(permissions);
  const editingRole = roles.find((r) => r.id === editingRoleId);
  const editingPerms = editingRoleId ? (rolePerms[editingRoleId] || []) : [];
  const assignedPermIds = new Set(editingPerms.map((p) => p.id));
  const availableToAdd = permissions.filter((p) => !assignedPermIds.has(p.id));

  return (
    <>
      <div className="page-header">
        <h1>Roles & Permissions</h1>
        <p>Flow 2 — Create roles, view predefined permissions, assign permissions to roles.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <PermTabBar tabs={visibleTabs} active={tab} onChange={setTab} />

      {tab === "roles" && (
        <>
          <PermGate perm="create_role">
            <div className="card">
              <div className="card-title">Create Role <span className="badge badge-post">POST</span></div>
              <form onSubmit={createRole}>
                <div className="grid-2">
                  <div className="form-group"><label>Name</label><input name="name" required value={roleForm.name} onChange={(e) => setRoleForm((p) => ({ ...p, name: e.target.value }))} placeholder="teacher" /></div>
                  <div className="form-group"><label>Description</label><input name="description" value={roleForm.description} onChange={(e) => setRoleForm((p) => ({ ...p, description: e.target.value }))} placeholder="Teacher role" /></div>
                </div>
                <div style={{ marginTop: 16 }}>
                  <div style={{ fontSize: 13, fontWeight: 600, marginBottom: 8 }}>Profile fields (shown when creating users with this role)</div>
                  {roleForm.fields?.length > 0 && (
                    <div style={{ display: "flex", flexWrap: "wrap", gap: 8, marginBottom: 12 }}>
                      {roleForm.fields.map((f, i) => (
                        <span key={i} className="perm-chip">
                          {f.label} ({f.key}){f.required ? " *" : ""}
                          <button type="button" title="Remove" onClick={() => removeField(i)}>&times;</button>
                        </span>
                      ))}
                    </div>
                  )}
                  <div className="grid-4">
                    <div className="form-group"><label>Field key</label><input value={newField.key} onChange={(e) => setNewField((p) => ({ ...p, key: e.target.value }))} placeholder="class_id" /></div>
                    <div className="form-group"><label>Label</label><input value={newField.label} onChange={(e) => setNewField((p) => ({ ...p, label: e.target.value }))} placeholder="Class" /></div>
                    <div className="form-group">
                      <label>Type</label>
                      <select value={newField.type} onChange={(e) => setNewField((p) => ({ ...p, type: e.target.value }))}>
                        {["text", "number", "email", "uuid", "select", "date"].map((t) => <option key={t} value={t}>{t}</option>)}
                      </select>
                    </div>
                    <div className="form-group" style={{ display: "flex", alignItems: "flex-end", gap: 8 }}>
                      <label><input type="checkbox" checked={newField.required} onChange={(e) => setNewField((p) => ({ ...p, required: e.target.checked }))} /> Required</label>
                      <button type="button" className="btn btn-ghost btn-sm" onClick={addFieldToRole}>Add field</button>
                    </div>
                  </div>
                </div>
                <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Create Role</button></div>
              </form>
            </div>
          </PermGate>
          <div className="card">
            <div className="card-title">Roles <span className="badge badge-get">GET</span></div>
            <div className="table-wrap">
              <table>
                <thead><tr><th>Name</th><th>Description</th><th>Permissions</th><th>Actions</th></tr></thead>
                <tbody>
                  {roles.length === 0 && <tr><td colSpan={4} className="empty">No roles.</td></tr>}
                  {roles.map((r) => (
                    <tr key={r.id}>
                      <td><strong>{r.name}</strong></td>
                      <td>{r.description}</td>
                      <td>
                        {rolePerms[r.id]?.length > 0 ? (
                          <span className="text-sm">{rolePerms[r.id].length} assigned</span>
                        ) : (
                          <span className="text-muted text-sm">—</span>
                        )}
                      </td>
                      <td>
                        <PermGate perm="manage_permissions">
                          <button type="button" className="btn btn-ghost btn-sm" onClick={() => openEditPerms(r)}>
                            Edit Permissions
                          </button>
                        </PermGate>
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
          {permissions.length === 0 && <div className="empty">No permissions found. Make sure auth-service is running.</div>}
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

      {editingRole && (
        <div className="modal-overlay" onClick={closeEditPerms}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Edit permissions — {editingRole.name}</h3>
              <button type="button" className="modal-close" onClick={closeEditPerms} aria-label="Close">&times;</button>
            </div>

            <p className="text-sm text-muted mb-4">Add or remove permissions for this role. Users with this role inherit these permissions.</p>

            <div style={{ display: "flex", flexWrap: "wrap", gap: 8, marginBottom: 16, minHeight: 32 }}>
              {editingPerms.length === 0 && (
                <em className="text-muted text-sm">No permissions assigned yet.</em>
              )}
              {editingPerms.map((p) => (
                <span key={p.id} className="perm-chip">
                  {p.name}
                  <button type="button" title="Remove" onClick={() => removePerm(editingRoleId, p.id)}>&times;</button>
                </span>
              ))}
            </div>

            <form onSubmit={addPermToRole}>
              <div className="form-group">
                <label>Add permission</label>
                <select
                  value={addPermId}
                  onChange={(e) => setAddPermId(e.target.value)}
                  disabled={availableToAdd.length === 0}
                >
                  <option value="">
                    {availableToAdd.length === 0 ? "All permissions already assigned" : "Select permission…"}
                  </option>
                  {availableToAdd.map((p) => (
                    <option key={p.id} value={p.id}>{p.name}</option>
                  ))}
                </select>
              </div>
              <div className="btn-row">
                <button type="button" className="btn btn-ghost" onClick={closeEditPerms}>Close</button>
                <button type="submit" className="btn btn-primary" disabled={busy || !addPermId}>Add Permission</button>
              </div>
            </form>
          </div>
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
