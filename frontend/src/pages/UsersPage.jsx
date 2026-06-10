import { useState, useEffect, useCallback } from "react";
import { userMgmtApi, rolesApi, academicApi } from "../api/client";
import PermGate from "../components/PermGate";
import ParentUserField from "../components/ParentUserField";
import { useAuth } from "../context/AuthContext";

const FIELD_TYPES = ["text", "number", "email", "uuid", "select", "date"];

export default function UsersPage() {
  const { hasPerm } = useAuth();
  const [users, setUsers] = useState([]);
  const [roles, setRoles] = useState([]);
  const [roleFields, setRoleFields] = useState([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({ page: 1, limit: 20, search: "", role_id: "" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);
  const [classes, setClasses] = useState([]);
  const [form, setForm] = useState({ name: "", email: "", password: "", role_id: "", role_data: {} });
  const [editingUser, setEditingUser] = useState(null);
  const [viewingUser, setViewingUser] = useState(null);
  const [viewRoleFields, setViewRoleFields] = useState([]);
  const [viewLoading, setViewLoading] = useState(false);
  const [editForm, setEditForm] = useState({ name: "", role_id: "", is_active: true, role_data: {} });
  const [editRoleFields, setEditRoleFields] = useState([]);

  const load = useCallback(async () => {
    try {
      const res = await userMgmtApi.list(query);
      setUsers(res.users || []);
      setTotal(res.total || 0);
    } catch (err) {
      setError(err.message);
    }
  }, [query]);

  useEffect(() => { load(); }, [load]);

  useEffect(() => {
    rolesApi.list().then(setRoles).catch(() => {});
    academicApi.getClasses().then(setClasses).catch(() => {});
  }, []);

  useEffect(() => {
    if (!form.role_id) {
      setRoleFields([]);
      return;
    }
    rolesApi.getFields(form.role_id).then(setRoleFields).catch(() => setRoleFields([]));
  }, [form.role_id]);

  useEffect(() => {
    if (!editForm.role_id) {
      setEditRoleFields([]);
      return;
    }
    rolesApi.getFields(editForm.role_id).then(setEditRoleFields).catch(() => setEditRoleFields([]));
  }, [editForm.role_id]);

  const flatSections = classes.flatMap((c) =>
    (c.sections || []).map((s) => ({ ...s, class_id: (c.class || c).id, className: (c.class || c).name }))
  );

  const flatClasses = classes.map((c) => c.class || c);
  const parentRole = roles.find((r) => r.name === "parent");
  const classNameById = Object.fromEntries(flatClasses.map((c) => [c.id, c.name]));
  const sectionNameById = Object.fromEntries(
    flatSections.map((s) => [s.id, `${s.className || ""} — ${s.name}`.trim()])
  );

  function field(e) {
    setForm((p) => ({ ...p, [e.target.name]: e.target.value }));
  }

  function roleDataField(key, value) {
    setForm((p) => ({ ...p, role_data: { ...p.role_data, [key]: value } }));
  }

  function selectParentForForm(parentId, parent) {
    setForm((p) => ({
      ...p,
      role_data: {
        ...p.role_data,
        parent_user_id: parentId || "",
        parent_name: parent?.name || "",
      },
    }));
  }

  function selectParentForEdit(parentId, parent) {
    setEditForm((p) => ({
      ...p,
      role_data: {
        ...p.role_data,
        parent_user_id: parentId || "",
        parent_name: parent?.name || "",
      },
    }));
  }

  function editRoleDataField(key, value) {
    setEditForm((p) => ({ ...p, role_data: { ...p.role_data, [key]: value } }));
  }

  function visibleRoleFields(fields) {
    return fields.filter((f) => f.key !== "parent_name");
  }

  function handleParentCreated(parent) {
    setSuccess(`Parent "${parent.name}" created and linked to this student.`);
    setQuery((q) => ({ ...q, page: 1, search: "" }));
  }

  async function handleCreate(e) {
    e.preventDefault();
    setError(""); setSuccess(""); setBusy(true);
    const studentRole = roles.find((r) => r.name === "student");
    if (studentRole && form.role_id === studentRole.id && !form.role_data?.parent_user_id) {
      setError("Please select or create a parent for this student.");
      setBusy(false);
      return;
    }
    try {
      await userMgmtApi.create(form);
      setSuccess("User created.");
      setForm({ name: "", email: "", password: "", role_id: "", role_data: {} });
      setRoleFields([]);
      load();
    } catch (err) { setError(err.message); }
    finally { setBusy(false); }
  }

  async function handleDelete(id) {
    if (!confirm("Delete this user?")) return;
    try {
      await userMgmtApi.remove(id);
      load();
    } catch (err) { setError(err.message); }
  }

  function openEdit(user) {
    setEditingUser(user);
    setEditForm({
      name: user.name || "",
      role_id: user.role_id || "",
      is_active: user.is_active !== false,
      role_data: user.role_data || {},
    });
    setError("");
    setSuccess("");
  }

  function closeEdit() {
    setEditingUser(null);
  }

  async function openView(user) {
    setViewLoading(true);
    setViewingUser(user);
    setViewRoleFields([]);
    setError("");
    try {
      const full = await userMgmtApi.getById(user.id);
      setViewingUser(full);
      if (full.role_id) {
        const fields = await rolesApi.getFields(full.role_id);
        setViewRoleFields(Array.isArray(fields) ? fields : []);
      }
    } catch (err) {
      setError(err.message);
      setViewingUser(null);
    } finally {
      setViewLoading(false);
    }
  }

  function closeView() {
    setViewingUser(null);
    setViewRoleFields([]);
  }

  function extraInfoSummary(roleData) {
    if (!roleData || Object.keys(roleData).length === 0) return null;
    if (roleData.student_code) return String(roleData.student_code);
    const filled = Object.entries(roleData).filter(([, v]) => v != null && String(v).trim() !== "");
    if (filled.length === 0) return null;
    return `${filled.length} field${filled.length > 1 ? "s" : ""}`;
  }

  function formatRoleFieldValue(key, value, fieldDef, roleData) {
    if (value == null || String(value).trim() === "") return "—";
    if (key === "class_id") return classNameById[String(value)] || String(value);
    if (key === "section_id") return sectionNameById[String(value)] || String(value);
    if (key === "parent_user_id") {
      const name = roleData?.parent_name;
      return name ? name : String(value);
    }
    if (fieldDef?.type === "uuid") return String(value);
    return String(value);
  }

  function roleDetailRows(user, fields) {
    const data = user?.role_data || {};
    const keys = fields?.length
      ? fields.map((f) => f.key)
      : Object.keys(data);
    const rows = [];
    const seen = new Set();
    for (const key of keys) {
      if (seen.has(key) || !(key in data)) continue;
      seen.add(key);
      const fieldDef = fields?.find((f) => f.key === key);
      rows.push({
        key,
        label: fieldDef?.label || key.replace(/_/g, " "),
        value: formatRoleFieldValue(key, data[key], fieldDef, data),
      });
    }
    for (const [key, value] of Object.entries(data)) {
      if (seen.has(key)) continue;
      rows.push({
        key,
        label: key.replace(/_/g, " "),
        value: formatRoleFieldValue(key, value, null, data),
      });
    }
    return rows;
  }

  async function handleUpdate(e) {
    e.preventDefault();
    if (!editingUser) return;
    setError("");
    setSuccess("");
    setBusy(true);
    try {
      await userMgmtApi.update(editingUser.id, {
        name: editForm.name,
        role_id: editForm.role_id || undefined,
        is_active: editForm.is_active,
        role_data: editForm.role_data,
      });
      setSuccess("User updated.");
      closeEdit();
      load();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  function renderRoleField(f, value, onChange, { mode = "create" } = {}) {
    if (f.key === "parent_user_id") {
      const onSelect = mode === "edit" ? selectParentForEdit : selectParentForForm;
      return (
        <ParentUserField
          required={f.required}
          parentRoleId={parentRole?.id}
          value={value || ""}
          onSelect={onSelect}
          onParentCreated={mode === "create" ? handleParentCreated : undefined}
        />
      );
    }
    if (f.key === "class_id") {
      const flatClasses = classes.map((c) => c.class || c);
      return (
        <select required={f.required} value={value || ""} onChange={(e) => onChange(f.key, e.target.value)}>
          <option value="">Select class…</option>
          {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
        </select>
      );
    }
    if (f.key === "section_id") {
      const classId = form.role_data?.class_id || editForm.role_data?.class_id;
      const sections = classId ? flatSections.filter((s) => s.class_id === classId) : [];
      return (
        <select required={f.required} value={value || ""} onChange={(e) => onChange(f.key, e.target.value)}>
          <option value="">Select section…</option>
          {sections.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
        </select>
      );
    }
    if (f.type === "select" && f.options?.length) {
      return (
        <select required={f.required} value={value || ""} onChange={(e) => onChange(f.key, e.target.value)}>
          <option value="">Select…</option>
          {f.options.map((o) => <option key={o} value={o}>{o}</option>)}
        </select>
      );
    }
    return (
      <input
        required={f.required}
        type={f.type === "number" ? "number" : f.type === "email" ? "email" : "text"}
        value={value || ""}
        onChange={(e) => onChange(f.key, e.target.value)}
        placeholder={f.label}
        readOnly={f.key === "student_code"}
      />
    );
  }

  return (
    <>
      <div className="page-header">
        <h1>User Management</h1>
        <p>
          Create and manage all users — staff, teachers, parents, and students.
          Select a role to see additional fields configured for that role.
        </p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <PermGate any={["create_user", "admit_student"]}>
      <div className="card">
        <div className="card-title">Create User <span className="badge badge-post">POST /users</span></div>
        <form onSubmit={handleCreate}>
          <div className="grid-4">
            <div className="form-group">
              <label>Full Name</label>
              <input name="name" required value={form.name} onChange={field} placeholder="Jane Smith" />
            </div>
            <div className="form-group">
              <label>Email</label>
              <input name="email" type="email" required value={form.email} onChange={field} placeholder="jane@school.edu" />
            </div>
            <div className="form-group">
              <label>Password</label>
              <input name="password" type="password" required minLength={6} value={form.password} onChange={field} placeholder="Min 6 chars" />
            </div>
            <div className="form-group">
              <label>Role</label>
              <select name="role_id" required value={form.role_id} onChange={field}>
                <option value="">Select role...</option>
                {roles.map((r) => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
            </div>
          </div>
          {visibleRoleFields(roleFields).length > 0 && (
            <div className="grid-4" style={{ marginTop: 16 }}>
              {visibleRoleFields(roleFields).map((f) => (
                <div className="form-group" key={f.key} style={f.key === "parent_user_id" ? { gridColumn: "1 / -1" } : undefined}>
                  <label>{f.label}{f.required ? " *" : ""}</label>
                  {renderRoleField(f, form.role_data[f.key], roleDataField, { mode: "create" })}
                </div>
              ))}
            </div>
          )}
          <div className="btn-row">
            <button className="btn btn-primary" disabled={busy}>{busy ? "Creating..." : "Create User"}</button>
          </div>
        </form>
      </div>
      </PermGate>

      <div className="card">
        <div className="card-title">
          Users ({total})
          <span className="badge badge-get">GET /users</span>
        </div>
        <div className="grid-2 mb-4">
          <div className="form-group">
            <label>Search</label>
            <input placeholder="Name or email..." value={query.search} onChange={(e) => setQuery((q) => ({ ...q, search: e.target.value, page: 1 }))} />
          </div>
          <div className="form-group">
            <label>Role</label>
            <select
              value={query.role_id || ""}
              onChange={(e) => setQuery((q) => ({ ...q, role_id: e.target.value, page: 1 }))}
            >
              <option value="">All roles</option>
              {roles.map((r) => (
                <option key={r.id} value={r.id}>{r.name}</option>
              ))}
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr><th>Name</th><th>Email</th><th>Role</th><th>Extra Info</th><th>Active</th><th>ID</th><th></th></tr>
            </thead>
            <tbody>
              {users.length === 0 && <tr><td colSpan={7} className="empty">No users found.</td></tr>}
              {users.map((u) => (
                <tr key={u.id}>
                  <td>{u.name}</td>
                  <td>{u.email}</td>
                  <td>{u.role_name || <span className="mono truncate">{u.role_id}</span>}</td>
                  <td className="text-sm">
                    <div className="btn-row" style={{ flexWrap: "nowrap", alignItems: "center", gap: 8 }}>
                      <span className="text-muted">{extraInfoSummary(u.role_data) || "—"}</span>
                      <button type="button" className="btn btn-ghost btn-sm" onClick={() => openView(u)}>
                        View
                      </button>
                    </div>
                  </td>
                  <td><span className={`status ${u.is_active ? "status-active" : "status-inactive"}`}>{u.is_active ? "Active" : "Inactive"}</span></td>
                  <td><span className="mono truncate">{u.id}</span></td>
                  <td>
                    <div className="btn-row" style={{ flexWrap: "nowrap" }}>
                      {hasPerm("update_user") && (
                        <button type="button" className="btn btn-ghost btn-sm" onClick={() => openEdit(u)}>Edit</button>
                      )}
                      {hasPerm("delete_user") && (
                        <button type="button" className="btn btn-danger btn-sm" onClick={() => handleDelete(u.id)}>Delete</button>
                      )}
                    </div>
                  </td>
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

      {viewingUser && (
        <div className="modal-overlay" onClick={closeView}>
          <div className="modal" style={{ maxWidth: 560 }} onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>User Details</h3>
              <button type="button" className="modal-close" onClick={closeView} aria-label="Close">&times;</button>
            </div>

            {viewLoading ? (
              <div className="empty">Loading…</div>
            ) : (
              <>
                <div style={{ marginBottom: 20 }}>
                  <div style={{ fontSize: 12, fontWeight: 700, textTransform: "uppercase", color: "var(--clr-text-secondary)", marginBottom: 8 }}>
                    Profile
                  </div>
                  <table className="detail-table" style={{ width: "100%", fontSize: 14 }}>
                    <tbody>
                      <tr><td style={{ width: "35%", color: "var(--clr-text-secondary)" }}>Name</td><td><strong>{viewingUser.name}</strong></td></tr>
                      <tr><td style={{ color: "var(--clr-text-secondary)" }}>Email</td><td>{viewingUser.email}</td></tr>
                      <tr><td style={{ color: "var(--clr-text-secondary)" }}>Role</td><td>{viewingUser.role_name || "—"}</td></tr>
                      <tr>
                        <td style={{ color: "var(--clr-text-secondary)" }}>Status</td>
                        <td>
                          <span className={`status ${viewingUser.is_active ? "status-active" : "status-inactive"}`}>
                            {viewingUser.is_active ? "Active" : "Inactive"}
                          </span>
                        </td>
                      </tr>
                      <tr><td style={{ color: "var(--clr-text-secondary)" }}>User ID</td><td><span className="mono">{viewingUser.id}</span></td></tr>
                    </tbody>
                  </table>
                </div>

                <div>
                  <div style={{ fontSize: 12, fontWeight: 700, textTransform: "uppercase", color: "var(--clr-text-secondary)", marginBottom: 8 }}>
                    Role-specific information
                  </div>
                  {roleDetailRows(viewingUser, viewRoleFields).length === 0 ? (
                    <div className="empty text-sm">No additional fields for this role.</div>
                  ) : (
                    <div className="table-wrap">
                      <table>
                        <thead>
                          <tr><th>Field</th><th>Value</th></tr>
                        </thead>
                        <tbody>
                          {roleDetailRows(viewingUser, viewRoleFields).map((row) => (
                            <tr key={row.key}>
                              <td>{row.label}</td>
                              <td>{row.value}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  )}
                </div>

                <div className="btn-row" style={{ marginTop: 20 }}>
                  <button type="button" className="btn btn-ghost" onClick={closeView}>Close</button>
                  {hasPerm("update_user") && (
                    <button
                      type="button"
                      className="btn btn-primary"
                      onClick={() => {
                        const u = viewingUser;
                        closeView();
                        openEdit(u);
                      }}
                    >
                      Edit user
                    </button>
                  )}
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {editingUser && hasPerm("update_user") && (
        <div className="modal-overlay" onClick={closeEdit}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Edit User</h3>
              <button type="button" className="modal-close" onClick={closeEdit} aria-label="Close">&times;</button>
            </div>
            <form onSubmit={handleUpdate}>
              <div className="form-group">
                <label>Full Name</label>
                <input required value={editForm.name} onChange={(e) => setEditForm((p) => ({ ...p, name: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Email</label>
                <input type="email" value={editingUser.email} disabled />
              </div>
              <div className="form-group">
                <label>Role</label>
                <select required value={editForm.role_id} onChange={(e) => setEditForm((p) => ({ ...p, role_id: e.target.value }))}>
                  <option value="">Select role…</option>
                  {roles.map((r) => (
                    <option key={r.id} value={r.id}>{r.name}</option>
                  ))}
                </select>
              </div>
              {visibleRoleFields(editRoleFields).length > 0 && visibleRoleFields(editRoleFields).map((f) => (
                <div className="form-group" key={f.key} style={f.key === "parent_user_id" ? { gridColumn: "1 / -1" } : undefined}>
                  <label>{f.label}{f.required ? " *" : ""}</label>
                  {renderRoleField(f, editForm.role_data[f.key], editRoleDataField, { mode: "edit" })}
                </div>
              ))}
              <div className="form-group">
                <label>
                  <input type="checkbox" checked={editForm.is_active} onChange={(e) => setEditForm((p) => ({ ...p, is_active: e.target.checked }))} />{" "}
                  Active
                </label>
              </div>
              <div className="btn-row">
                <button type="button" className="btn btn-ghost" onClick={closeEdit}>Cancel</button>
                <button type="submit" className="btn btn-primary" disabled={busy}>{busy ? "Saving…" : "Save"}</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </>
  );
}

export { FIELD_TYPES };
