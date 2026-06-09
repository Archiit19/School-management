import { useState, useEffect, useCallback } from "react";
import { userMgmtApi, rolesApi } from "../api/client";
import PermGate from "../components/PermGate";
import { useAuth } from "../context/AuthContext";

export default function UsersPage() {
  const { hasPerm } = useAuth();
  const [users, setUsers] = useState([]);
  const [roles, setRoles] = useState([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({ page: 1, limit: 20, search: "" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);
  const [form, setForm] = useState({ name: "", email: "", password: "", role_id: "" });
  const [editingUser, setEditingUser] = useState(null);
  const [editForm, setEditForm] = useState({ name: "", role_id: "", is_active: true });

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
  }, []);

  function field(e) {
    setForm((p) => ({ ...p, [e.target.name]: e.target.value }));
  }

  async function handleCreate(e) {
    e.preventDefault();
    setError(""); setSuccess(""); setBusy(true);
    try {
      await userMgmtApi.create(form);
      setSuccess("User created.");
      setForm({ name: "", email: "", password: "", role_id: "" });
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
    });
    setError("");
    setSuccess("");
  }

  function closeEdit() {
    setEditingUser(null);
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

  const editableRoles = roles.filter((r) => r.name && r.name.toLowerCase() !== "student");

  return (
    <>
      <div className="page-header">
        <h1>User Management</h1>
        <p>
          Create and manage staff, teachers, and parents. Pupils are enrolled under{" "}
          <strong>Students</strong> (admit students), not here — the Student role is not offered below.
          Requires permission to manage users.
        </p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <PermGate perm="create_user">
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
                {roles
                  .filter((r) => r.name && r.name.toLowerCase() !== "student")
                  .map((r) => (
                    <option key={r.id} value={r.id}>{r.name}</option>
                  ))}
              </select>
            </div>
          </div>
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
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr><th>Name</th><th>Email</th><th>Role</th><th>Active</th><th>ID</th><th></th></tr>
            </thead>
            <tbody>
              {users.length === 0 && <tr><td colSpan={6} className="empty">No users found.</td></tr>}
              {users.map((u) => (
                <tr key={u.id}>
                  <td>{u.name}</td>
                  <td>{u.email}</td>
                  <td>{u.role_name || <span className="mono truncate">{u.role_id}</span>}</td>
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
                <input
                  required
                  value={editForm.name}
                  onChange={(e) => setEditForm((p) => ({ ...p, name: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>Email</label>
                <input type="email" value={editingUser.email} disabled />
              </div>
              <div className="form-group">
                <label>Role</label>
                <select
                  required
                  value={editForm.role_id}
                  onChange={(e) => setEditForm((p) => ({ ...p, role_id: e.target.value }))}
                >
                  <option value="">Select role…</option>
                  {editableRoles.map((r) => (
                    <option key={r.id} value={r.id}>{r.name}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>
                  <input
                    type="checkbox"
                    checked={editForm.is_active}
                    onChange={(e) => setEditForm((p) => ({ ...p, is_active: e.target.checked }))}
                  />{" "}
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
