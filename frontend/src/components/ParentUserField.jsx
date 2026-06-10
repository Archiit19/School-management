import { useState, useEffect, useRef } from "react";
import { userMgmtApi, rolesApi } from "../api/client";

const ADD_NEW = "__add_new__";

export default function ParentUserField({
  value,
  onSelect,
  onParentCreated,
  parentRoleId,
  required = false,
  disabled = false,
}) {
  const [search, setSearch] = useState("");
  const [parents, setParents] = useState([]);
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [selectedLabel, setSelectedLabel] = useState("");
  const [showAddForm, setShowAddForm] = useState(false);
  const [parentRoleFields, setParentRoleFields] = useState([]);
  const [addForm, setAddForm] = useState({
    name: "",
    email: "",
    password: "",
    role_data: {},
  });
  const [addBusy, setAddBusy] = useState(false);
  const [addError, setAddError] = useState("");
  const wrapRef = useRef(null);

  useEffect(() => {
    if (!parentRoleId) return;
    rolesApi.getFields(parentRoleId).then(setParentRoleFields).catch(() => setParentRoleFields([]));
  }, [parentRoleId]);

  useEffect(() => {
    if (!value) {
      setSelectedLabel("");
      return;
    }
    let cancelled = false;
    userMgmtApi
      .getById(value)
      .then((u) => {
        if (!cancelled) setSelectedLabel(`${u.name} (${u.email})`);
      })
      .catch(() => {
        if (!cancelled) setSelectedLabel(String(value));
      });
    return () => {
      cancelled = true;
    };
  }, [value]);

  useEffect(() => {
    if (!parentRoleId || !open) return undefined;
    const timer = setTimeout(() => {
      setLoading(true);
      userMgmtApi
        .list({ role_id: parentRoleId, search, limit: 50, page: 1 })
        .then((res) => setParents(res.users || []))
        .catch(() => setParents([]))
        .finally(() => setLoading(false));
    }, 250);
    return () => clearTimeout(timer);
  }, [parentRoleId, search, open]);

  useEffect(() => {
    function onDocClick(e) {
      if (wrapRef.current && !wrapRef.current.contains(e.target)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", onDocClick);
    return () => document.removeEventListener("mousedown", onDocClick);
  }, []);

  function pickParent(parent) {
    onSelect(parent.id, parent);
    setSelectedLabel(`${parent.name} (${parent.email})`);
    setShowAddForm(false);
    setOpen(false);
    setSearch("");
  }

  function handlePick(value) {
    if (value === ADD_NEW) {
      setShowAddForm(true);
      setOpen(false);
      return;
    }
    const parent = parents.find((p) => p.id === value);
    if (parent) pickParent(parent);
  }

  async function handleCreateParent(e) {
    if (e?.preventDefault) e.preventDefault();
    if (e?.stopPropagation) e.stopPropagation();
    if (!parentRoleId) return;
    setAddError("");
    setAddBusy(true);
    try {
      const created = await userMgmtApi.create({
        name: addForm.name,
        email: addForm.email,
        password: addForm.password,
        role_id: parentRoleId,
        role_data: addForm.role_data,
      });
      setParents((prev) => [created, ...prev.filter((p) => p.id !== created.id)]);
      pickParent(created);
      onParentCreated?.(created);
      setAddForm({ name: "", email: "", password: "", role_data: {} });
      setShowAddForm(false);
    } catch (err) {
      setAddError(err.message);
    } finally {
      setAddBusy(false);
    }
  }

  function parentDataField(key, val) {
    setAddForm((p) => ({ ...p, role_data: { ...p.role_data, [key]: val } }));
  }

  if (!parentRoleId) {
    return <div className="text-sm text-muted">Parent role not configured for this school.</div>;
  }

  return (
    <div className="parent-picker" ref={wrapRef}>
      <div className="parent-picker-control">
        <input
          type="text"
          required={required && !value}
          disabled={disabled}
          placeholder={value ? selectedLabel : "Search parents by name or email…"}
          value={open ? search : selectedLabel || ""}
          onChange={(e) => {
            setSearch(e.target.value);
            if (!open) setOpen(true);
          }}
          onFocus={() => setOpen(true)}
          readOnly={!!value && !open}
          onClick={() => {
            if (value && !open) {
              setOpen(true);
              setSearch("");
            }
          }}
        />
        {value && (
          <button
            type="button"
            className="btn btn-ghost btn-sm parent-picker-clear"
            onClick={() => {
              onSelect("", null);
              setSelectedLabel("");
              setSearch("");
              setShowAddForm(false);
            }}
            aria-label="Clear parent"
          >
            Clear
          </button>
        )}
      </div>

      {open && (
        <div className="parent-picker-dropdown">
          <button
            type="button"
            className="parent-picker-option parent-picker-option-new"
            onClick={() => handlePick(ADD_NEW)}
          >
            + Add new parent
          </button>
          {loading && <div className="parent-picker-empty">Loading…</div>}
          {!loading && parents.length === 0 && (
            <div className="parent-picker-empty">No parents match your search.</div>
          )}
          {!loading &&
            parents.map((p) => (
              <button
                type="button"
                key={p.id}
                className={`parent-picker-option${p.id === value ? " active" : ""}`}
                onClick={() => handlePick(p.id)}
              >
                <strong>{p.name}</strong>
                <span className="text-muted">{p.email}</span>
              </button>
            ))}
        </div>
      )}

      {showAddForm && (
        <div className="parent-picker-add card" style={{ marginTop: 12, padding: 16 }}>
          <div style={{ fontWeight: 600, marginBottom: 12 }}>Create parent account</div>
          {addError && <div className="alert alert-error" style={{ marginBottom: 12 }}>{addError}</div>}
          <div
            onKeyDown={(e) => {
              if (e.key === "Enter" && e.target.tagName !== "TEXTAREA") {
                e.preventDefault();
                e.stopPropagation();
                handleCreateParent(e);
              }
            }}
          >
            <div className="grid-2">
              <div className="form-group">
                <label>Full Name *</label>
                <input
                  required
                  value={addForm.name}
                  onChange={(e) => setAddForm((p) => ({ ...p, name: e.target.value }))}
                  placeholder="Parent name"
                />
              </div>
              <div className="form-group">
                <label>Email *</label>
                <input
                  type="email"
                  required
                  value={addForm.email}
                  onChange={(e) => setAddForm((p) => ({ ...p, email: e.target.value }))}
                  placeholder="parent@email.com"
                />
              </div>
              <div className="form-group">
                <label>Password *</label>
                <input
                  type="password"
                  required
                  minLength={6}
                  value={addForm.password}
                  onChange={(e) => setAddForm((p) => ({ ...p, password: e.target.value }))}
                  placeholder="Min 6 characters"
                />
              </div>
              {parentRoleFields.map((f) => (
                <div className="form-group" key={f.key}>
                  <label>{f.label}{f.required ? " *" : ""}</label>
                  <input
                    required={f.required}
                    value={addForm.role_data[f.key] || ""}
                    onChange={(e) => parentDataField(f.key, e.target.value)}
                    placeholder={f.label}
                  />
                </div>
              ))}
            </div>
            <div className="btn-row" style={{ marginTop: 8 }}>
              <button
                type="button"
                className="btn btn-primary btn-sm"
                disabled={addBusy}
                onClick={handleCreateParent}
              >
                {addBusy ? "Creating…" : "Create & select parent"}
              </button>
              <button
                type="button"
                className="btn btn-ghost btn-sm"
                onClick={() => {
                  setShowAddForm(false);
                  setAddError("");
                }}
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
