import { useState, useEffect, useCallback } from "react";
import { transportApi, studentApi } from "../api/client";
import { useAuth } from "../context/AuthContext";

const VEHICLE_TYPES = ["bus", "van", "mini-bus"];
const TRANSPORT_TYPES = ["pickup", "drop", "both"];

export default function TransportPage() {
  const { hasPerm } = useAuth();
  const canManage = hasPerm("manage_transport");
  const canView = hasPerm("view_transport") || canManage;

  const [tab, setTab] = useState("vehicles");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  if (!canView) {
    return <div className="card"><p>You don't have permission to view transport information.</p></div>;
  }

  return (
    <>
      <div className="page-header">
        <h1>Transport Management</h1>
        <p>Manage vehicles, routes, stops, and student transport assignments.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="tabs">
        <button className={`tab ${tab === "vehicles" ? "active" : ""}`} onClick={() => { setTab("vehicles"); setError(""); setSuccess(""); }}>
          Vehicles
        </button>
        <button className={`tab ${tab === "routes" ? "active" : ""}`} onClick={() => { setTab("routes"); setError(""); setSuccess(""); }}>
          Routes
        </button>
        <button className={`tab ${tab === "stops" ? "active" : ""}`} onClick={() => { setTab("stops"); setError(""); setSuccess(""); }}>
          Stops
        </button>
        <button className={`tab ${tab === "assignments" ? "active" : ""}`} onClick={() => { setTab("assignments"); setError(""); setSuccess(""); }}>
          Student Assignments
        </button>
      </div>

      {tab === "vehicles" && <VehiclesTab canManage={canManage} setError={setError} setSuccess={setSuccess} />}
      {tab === "routes" && <RoutesTab canManage={canManage} setError={setError} setSuccess={setSuccess} />}
      {tab === "stops" && <StopsTab canManage={canManage} setError={setError} setSuccess={setSuccess} />}
      {tab === "assignments" && <AssignmentsTab canManage={canManage} setError={setError} setSuccess={setSuccess} />}
    </>
  );
}

function VehiclesTab({ canManage, setError, setSuccess }) {
  const [vehicles, setVehicles] = useState([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({ page: 1, limit: 20, vehicle_type: "", is_active: "" });
  const [showForm, setShowForm] = useState(false);
  const [editId, setEditId] = useState(null);
  const [form, setForm] = useState({
    vehicle_number: "", vehicle_type: "bus", capacity: 40,
    driver_name: "", driver_phone: "", driver_license: "",
    conductor_name: "", conductor_phone: "",
    insurance_expiry: "", fitness_expiry: ""
  });
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    try {
      const res = await transportApi.getVehicles(query);
      setVehicles(res.vehicles || []);
      setTotal(res.total || 0);
    } catch (err) {
      setError(err.message);
    }
  }, [query, setError]);

  useEffect(() => { load(); }, [load]);

  async function handleSubmit(e) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const payload = { ...form, capacity: parseInt(form.capacity, 10) };
      if (editId) {
        await transportApi.updateVehicle(editId, payload);
        setSuccess("Vehicle updated.");
      } else {
        await transportApi.createVehicle(payload);
        setSuccess("Vehicle created.");
      }
      setTimeout(() => setSuccess(""), 3000);
      setShowForm(false);
      setEditId(null);
      resetForm();
      load();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  function resetForm() {
    setForm({
      vehicle_number: "", vehicle_type: "bus", capacity: 40,
      driver_name: "", driver_phone: "", driver_license: "",
      conductor_name: "", conductor_phone: "",
      insurance_expiry: "", fitness_expiry: ""
    });
  }

  function startEdit(v) {
    setEditId(v.id);
    setForm({
      vehicle_number: v.vehicle_number,
      vehicle_type: v.vehicle_type,
      capacity: v.capacity,
      driver_name: v.driver_name || "",
      driver_phone: v.driver_phone || "",
      driver_license: v.driver_license || "",
      conductor_name: v.conductor_name || "",
      conductor_phone: v.conductor_phone || "",
      insurance_expiry: v.insurance_expiry?.split("T")[0] || "",
      fitness_expiry: v.fitness_expiry?.split("T")[0] || ""
    });
    setShowForm(true);
  }

  return (
    <>
      {canManage && (
        <div className="btn-row mb-4">
          <button className="btn btn-primary" onClick={() => { setShowForm(!showForm); setEditId(null); resetForm(); }}>
            {showForm ? "Cancel" : "+ Add Vehicle"}
          </button>
        </div>
      )}

      {showForm && canManage && (
        <div className="card mb-4">
          <div className="card-title">{editId ? "Edit Vehicle" : "Add Vehicle"}</div>
          <form onSubmit={handleSubmit}>
            <div className="grid-4">
              <div className="form-group">
                <label>Vehicle Number *</label>
                <input required value={form.vehicle_number} onChange={(e) => setForm(f => ({ ...f, vehicle_number: e.target.value }))} placeholder="MH12AB1234" />
              </div>
              <div className="form-group">
                <label>Type *</label>
                <select value={form.vehicle_type} onChange={(e) => setForm(f => ({ ...f, vehicle_type: e.target.value }))}>
                  {VEHICLE_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Capacity *</label>
                <input type="number" required min="1" value={form.capacity} onChange={(e) => setForm(f => ({ ...f, capacity: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Driver Name</label>
                <input value={form.driver_name} onChange={(e) => setForm(f => ({ ...f, driver_name: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Driver Phone</label>
                <input value={form.driver_phone} onChange={(e) => setForm(f => ({ ...f, driver_phone: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Driver License</label>
                <input value={form.driver_license} onChange={(e) => setForm(f => ({ ...f, driver_license: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Conductor Name</label>
                <input value={form.conductor_name} onChange={(e) => setForm(f => ({ ...f, conductor_name: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Conductor Phone</label>
                <input value={form.conductor_phone} onChange={(e) => setForm(f => ({ ...f, conductor_phone: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Insurance Expiry</label>
                <input type="date" value={form.insurance_expiry} onChange={(e) => setForm(f => ({ ...f, insurance_expiry: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Fitness Expiry</label>
                <input type="date" value={form.fitness_expiry} onChange={(e) => setForm(f => ({ ...f, fitness_expiry: e.target.value }))} />
              </div>
            </div>
            <div className="btn-row">
              <button className="btn btn-primary" disabled={busy}>{busy ? "Saving..." : editId ? "Update" : "Create"}</button>
            </div>
          </form>
        </div>
      )}

      <div className="card">
        <div className="card-title">Vehicles ({total})</div>
        <div className="grid-3 mb-4">
          <div className="form-group">
            <label>Type</label>
            <select value={query.vehicle_type} onChange={(e) => setQuery(q => ({ ...q, vehicle_type: e.target.value, page: 1 }))}>
              <option value="">All</option>
              {VEHICLE_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
            </select>
          </div>
          <div className="form-group">
            <label>Status</label>
            <select value={query.is_active} onChange={(e) => setQuery(q => ({ ...q, is_active: e.target.value, page: 1 }))}>
              <option value="">All</option>
              <option value="true">Active</option>
              <option value="false">Inactive</option>
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Number</th>
                <th>Type</th>
                <th>Capacity</th>
                <th>Driver</th>
                <th>Phone</th>
                <th>Status</th>
                {canManage && <th></th>}
              </tr>
            </thead>
            <tbody>
              {vehicles.length === 0 && <tr><td colSpan={canManage ? 7 : 6} className="empty">No vehicles found.</td></tr>}
              {vehicles.map(v => (
                <tr key={v.id}>
                  <td><strong>{v.vehicle_number}</strong></td>
                  <td>{v.vehicle_type}</td>
                  <td>{v.capacity}</td>
                  <td>{v.driver_name || "-"}</td>
                  <td>{v.driver_phone || "-"}</td>
                  <td><span className={`status status-${v.is_active ? "present" : "absent"}`}>{v.is_active ? "Active" : "Inactive"}</span></td>
                  {canManage && <td><button className="btn btn-ghost btn-sm" onClick={() => startEdit(v)}>Edit</button></td>}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

function RoutesTab({ canManage, setError, setSuccess }) {
  const [routes, setRoutes] = useState([]);
  const [vehicles, setVehicles] = useState([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({ page: 1, limit: 20, is_active: "" });
  const [showForm, setShowForm] = useState(false);
  const [editId, setEditId] = useState(null);
  const [form, setForm] = useState({
    route_name: "", route_code: "", vehicle_id: "",
    start_point: "", end_point: "", distance: "", duration: "", monthly_fee: ""
  });
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    try {
      const [routeRes, vehicleRes] = await Promise.all([
        transportApi.getRoutes(query),
        transportApi.getVehicles({ limit: 100, is_active: "true" })
      ]);
      setRoutes(routeRes.routes || []);
      setTotal(routeRes.total || 0);
      setVehicles(vehicleRes.vehicles || []);
    } catch (err) {
      setError(err.message);
    }
  }, [query, setError]);

  useEffect(() => { load(); }, [load]);

  async function handleSubmit(e) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const payload = {
        ...form,
        distance: form.distance ? parseFloat(form.distance) : 0,
        duration: form.duration ? parseInt(form.duration, 10) : 0,
        monthly_fee: form.monthly_fee ? parseFloat(form.monthly_fee) : 0
      };
      if (editId) {
        await transportApi.updateRoute(editId, payload);
        setSuccess("Route updated.");
      } else {
        await transportApi.createRoute(payload);
        setSuccess("Route created.");
      }
      setTimeout(() => setSuccess(""), 3000);
      setShowForm(false);
      setEditId(null);
      resetForm();
      load();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  function resetForm() {
    setForm({ route_name: "", route_code: "", vehicle_id: "", start_point: "", end_point: "", distance: "", duration: "", monthly_fee: "" });
  }

  function startEdit(r) {
    setEditId(r.id);
    setForm({
      route_name: r.route_name,
      route_code: r.route_code,
      vehicle_id: r.vehicle_id || "",
      start_point: r.start_point || "",
      end_point: r.end_point || "",
      distance: r.distance || "",
      duration: r.duration || "",
      monthly_fee: r.monthly_fee || ""
    });
    setShowForm(true);
  }

  return (
    <>
      {canManage && (
        <div className="btn-row mb-4">
          <button className="btn btn-primary" onClick={() => { setShowForm(!showForm); setEditId(null); resetForm(); }}>
            {showForm ? "Cancel" : "+ Add Route"}
          </button>
        </div>
      )}

      {showForm && canManage && (
        <div className="card mb-4">
          <div className="card-title">{editId ? "Edit Route" : "Add Route"}</div>
          <form onSubmit={handleSubmit}>
            <div className="grid-4">
              <div className="form-group">
                <label>Route Name *</label>
                <input required value={form.route_name} onChange={(e) => setForm(f => ({ ...f, route_name: e.target.value }))} placeholder="North Zone" />
              </div>
              <div className="form-group">
                <label>Route Code *</label>
                <input required value={form.route_code} onChange={(e) => setForm(f => ({ ...f, route_code: e.target.value }))} placeholder="R001" />
              </div>
              <div className="form-group">
                <label>Vehicle</label>
                <select value={form.vehicle_id} onChange={(e) => setForm(f => ({ ...f, vehicle_id: e.target.value }))}>
                  <option value="">None</option>
                  {vehicles.map(v => <option key={v.id} value={v.id}>{v.vehicle_number} ({v.vehicle_type})</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Start Point</label>
                <input value={form.start_point} onChange={(e) => setForm(f => ({ ...f, start_point: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>End Point</label>
                <input value={form.end_point} onChange={(e) => setForm(f => ({ ...f, end_point: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Distance (km)</label>
                <input type="number" step="0.1" value={form.distance} onChange={(e) => setForm(f => ({ ...f, distance: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Duration (min)</label>
                <input type="number" value={form.duration} onChange={(e) => setForm(f => ({ ...f, duration: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Monthly Fee</label>
                <input type="number" step="0.01" value={form.monthly_fee} onChange={(e) => setForm(f => ({ ...f, monthly_fee: e.target.value }))} />
              </div>
            </div>
            <div className="btn-row">
              <button className="btn btn-primary" disabled={busy}>{busy ? "Saving..." : editId ? "Update" : "Create"}</button>
            </div>
          </form>
        </div>
      )}

      <div className="card">
        <div className="card-title">Routes ({total})</div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Code</th>
                <th>Name</th>
                <th>Start - End</th>
                <th>Distance</th>
                <th>Duration</th>
                <th>Fee</th>
                <th>Status</th>
                {canManage && <th></th>}
              </tr>
            </thead>
            <tbody>
              {routes.length === 0 && <tr><td colSpan={canManage ? 8 : 7} className="empty">No routes found.</td></tr>}
              {routes.map(r => (
                <tr key={r.id}>
                  <td><strong>{r.route_code}</strong></td>
                  <td>{r.route_name}</td>
                  <td>{r.start_point || "-"} → {r.end_point || "-"}</td>
                  <td>{r.distance ? `${r.distance} km` : "-"}</td>
                  <td>{r.duration ? `${r.duration} min` : "-"}</td>
                  <td>{r.monthly_fee ? `₹${r.monthly_fee}` : "-"}</td>
                  <td><span className={`status status-${r.is_active ? "present" : "absent"}`}>{r.is_active ? "Active" : "Inactive"}</span></td>
                  {canManage && <td><button className="btn btn-ghost btn-sm" onClick={() => startEdit(r)}>Edit</button></td>}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

function StopsTab({ canManage, setError, setSuccess }) {
  const [stops, setStops] = useState([]);
  const [routes, setRoutes] = useState([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({ page: 1, limit: 50, route_id: "" });
  const [showForm, setShowForm] = useState(false);
  const [editId, setEditId] = useState(null);
  const [form, setForm] = useState({
    route_id: "", stop_name: "", stop_order: 1, pickup_time: "", drop_time: "", landmark: ""
  });
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    try {
      const [stopRes, routeRes] = await Promise.all([
        transportApi.getStops(query),
        transportApi.getRoutes({ limit: 100, is_active: "true" })
      ]);
      setStops(stopRes.stops || []);
      setTotal(stopRes.total || 0);
      setRoutes(routeRes.routes || []);
    } catch (err) {
      setError(err.message);
    }
  }, [query, setError]);

  useEffect(() => { load(); }, [load]);

  async function handleSubmit(e) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const payload = { ...form, stop_order: parseInt(form.stop_order, 10) };
      if (editId) {
        await transportApi.updateStop(editId, payload);
        setSuccess("Stop updated.");
      } else {
        await transportApi.createStop(payload);
        setSuccess("Stop created.");
      }
      setTimeout(() => setSuccess(""), 3000);
      setShowForm(false);
      setEditId(null);
      resetForm();
      load();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  function resetForm() {
    setForm({ route_id: "", stop_name: "", stop_order: 1, pickup_time: "", drop_time: "", landmark: "" });
  }

  function startEdit(s) {
    setEditId(s.id);
    setForm({
      route_id: s.route_id,
      stop_name: s.stop_name,
      stop_order: s.stop_order,
      pickup_time: s.pickup_time || "",
      drop_time: s.drop_time || "",
      landmark: s.landmark || ""
    });
    setShowForm(true);
  }

  const routeMap = Object.fromEntries(routes.map(r => [r.id, r]));

  return (
    <>
      {canManage && (
        <div className="btn-row mb-4">
          <button className="btn btn-primary" onClick={() => { setShowForm(!showForm); setEditId(null); resetForm(); }}>
            {showForm ? "Cancel" : "+ Add Stop"}
          </button>
        </div>
      )}

      {showForm && canManage && (
        <div className="card mb-4">
          <div className="card-title">{editId ? "Edit Stop" : "Add Stop"}</div>
          <form onSubmit={handleSubmit}>
            <div className="grid-3">
              <div className="form-group">
                <label>Route *</label>
                <select required value={form.route_id} onChange={(e) => setForm(f => ({ ...f, route_id: e.target.value }))} disabled={!!editId}>
                  <option value="">Select route...</option>
                  {routes.map(r => <option key={r.id} value={r.id}>{r.route_code} - {r.route_name}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Stop Name *</label>
                <input required value={form.stop_name} onChange={(e) => setForm(f => ({ ...f, stop_name: e.target.value }))} placeholder="Main Market" />
              </div>
              <div className="form-group">
                <label>Order *</label>
                <input type="number" required min="1" value={form.stop_order} onChange={(e) => setForm(f => ({ ...f, stop_order: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Pickup Time</label>
                <input type="time" value={form.pickup_time} onChange={(e) => setForm(f => ({ ...f, pickup_time: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Drop Time</label>
                <input type="time" value={form.drop_time} onChange={(e) => setForm(f => ({ ...f, drop_time: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>Landmark</label>
                <input value={form.landmark} onChange={(e) => setForm(f => ({ ...f, landmark: e.target.value }))} placeholder="Near temple" />
              </div>
            </div>
            <div className="btn-row">
              <button className="btn btn-primary" disabled={busy}>{busy ? "Saving..." : editId ? "Update" : "Create"}</button>
            </div>
          </form>
        </div>
      )}

      <div className="card">
        <div className="card-title">Stops ({total})</div>
        <div className="grid-2 mb-4">
          <div className="form-group">
            <label>Filter by Route</label>
            <select value={query.route_id} onChange={(e) => setQuery(q => ({ ...q, route_id: e.target.value, page: 1 }))}>
              <option value="">All Routes</option>
              {routes.map(r => <option key={r.id} value={r.id}>{r.route_code} - {r.route_name}</option>)}
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Order</th>
                <th>Stop Name</th>
                <th>Route</th>
                <th>Pickup</th>
                <th>Drop</th>
                <th>Landmark</th>
                {canManage && <th></th>}
              </tr>
            </thead>
            <tbody>
              {stops.length === 0 && <tr><td colSpan={canManage ? 7 : 6} className="empty">No stops found.</td></tr>}
              {stops.map(s => (
                <tr key={s.id}>
                  <td><strong>{s.stop_order}</strong></td>
                  <td>{s.stop_name}</td>
                  <td>{routeMap[s.route_id]?.route_code || s.route_id}</td>
                  <td>{s.pickup_time || "-"}</td>
                  <td>{s.drop_time || "-"}</td>
                  <td>{s.landmark || "-"}</td>
                  {canManage && <td><button className="btn btn-ghost btn-sm" onClick={() => startEdit(s)}>Edit</button></td>}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}

function AssignmentsTab({ canManage, setError, setSuccess }) {
  const [assignments, setAssignments] = useState([]);
  const [routes, setRoutes] = useState([]);
  const [stops, setStops] = useState([]);
  const [total, setTotal] = useState(0);
  const [query, setQuery] = useState({ page: 1, limit: 20, route_id: "", is_active: "" });
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({
    student_id: "", route_id: "", stop_id: "", transport_type: "both", start_date: new Date().toISOString().split("T")[0], end_date: ""
  });
  const [busy, setBusy] = useState(false);

  const load = useCallback(async () => {
    try {
      const [assignRes, routeRes, stopRes] = await Promise.all([
        transportApi.getStudentTransports(query),
        transportApi.getRoutes({ limit: 100, is_active: "true" }),
        transportApi.getStops({ limit: 200 })
      ]);
      setAssignments(assignRes.assignments || []);
      setTotal(assignRes.total || 0);
      setRoutes(routeRes.routes || []);
      setStops(stopRes.stops || []);
    } catch (err) {
      setError(err.message);
    }
  }, [query, setError]);

  useEffect(() => { load(); }, [load]);

  async function handleSubmit(e) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await transportApi.assignStudent(form);
      setSuccess("Student assigned to transport.");
      setTimeout(() => setSuccess(""), 3000);
      setShowForm(false);
      resetForm();
      load();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  function resetForm() {
    setForm({ student_id: "", route_id: "", stop_id: "", transport_type: "both", start_date: new Date().toISOString().split("T")[0], end_date: "" });
  }

  const routeMap = Object.fromEntries(routes.map(r => [r.id, r]));
  const stopMap = Object.fromEntries(stops.map(s => [s.id, s]));
  const filteredStops = form.route_id ? stops.filter(s => s.route_id === form.route_id) : stops;

  return (
    <>
      {canManage && (
        <div className="btn-row mb-4">
          <button className="btn btn-primary" onClick={() => { setShowForm(!showForm); resetForm(); }}>
            {showForm ? "Cancel" : "+ Assign Student"}
          </button>
        </div>
      )}

      {showForm && canManage && (
        <div className="card mb-4">
          <div className="card-title">Assign Student to Transport</div>
          <form onSubmit={handleSubmit}>
            <div className="grid-3">
              <div className="form-group">
                <label>Student ID *</label>
                <input required value={form.student_id} onChange={(e) => setForm(f => ({ ...f, student_id: e.target.value }))} placeholder="Student UUID" />
              </div>
              <div className="form-group">
                <label>Route *</label>
                <select required value={form.route_id} onChange={(e) => setForm(f => ({ ...f, route_id: e.target.value, stop_id: "" }))}>
                  <option value="">Select route...</option>
                  {routes.map(r => <option key={r.id} value={r.id}>{r.route_code} - {r.route_name}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Stop *</label>
                <select required value={form.stop_id} onChange={(e) => setForm(f => ({ ...f, stop_id: e.target.value }))}>
                  <option value="">Select stop...</option>
                  {filteredStops.map(s => <option key={s.id} value={s.id}>{s.stop_order}. {s.stop_name}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Type *</label>
                <select value={form.transport_type} onChange={(e) => setForm(f => ({ ...f, transport_type: e.target.value }))}>
                  {TRANSPORT_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label>Start Date *</label>
                <input type="date" required value={form.start_date} onChange={(e) => setForm(f => ({ ...f, start_date: e.target.value }))} />
              </div>
              <div className="form-group">
                <label>End Date</label>
                <input type="date" value={form.end_date} onChange={(e) => setForm(f => ({ ...f, end_date: e.target.value }))} />
              </div>
            </div>
            <div className="btn-row">
              <button className="btn btn-primary" disabled={busy}>{busy ? "Assigning..." : "Assign"}</button>
            </div>
          </form>
        </div>
      )}

      <div className="card">
        <div className="card-title">Student Transport Assignments ({total})</div>
        <div className="grid-2 mb-4">
          <div className="form-group">
            <label>Filter by Route</label>
            <select value={query.route_id} onChange={(e) => setQuery(q => ({ ...q, route_id: e.target.value, page: 1 }))}>
              <option value="">All Routes</option>
              {routes.map(r => <option key={r.id} value={r.id}>{r.route_code} - {r.route_name}</option>)}
            </select>
          </div>
          <div className="form-group">
            <label>Status</label>
            <select value={query.is_active} onChange={(e) => setQuery(q => ({ ...q, is_active: e.target.value, page: 1 }))}>
              <option value="">All</option>
              <option value="true">Active</option>
              <option value="false">Inactive</option>
            </select>
          </div>
        </div>
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Student ID</th>
                <th>Route</th>
                <th>Stop</th>
                <th>Type</th>
                <th>Start</th>
                <th>End</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {assignments.length === 0 && <tr><td colSpan={7} className="empty">No assignments found.</td></tr>}
              {assignments.map(a => (
                <tr key={a.id}>
                  <td><span className="mono truncate">{a.student_id}</span></td>
                  <td>{routeMap[a.route_id]?.route_code || a.route_id}</td>
                  <td>{stopMap[a.stop_id]?.stop_name || a.stop_id}</td>
                  <td>{a.transport_type}</td>
                  <td>{a.start_date?.split("T")[0]}</td>
                  <td>{a.end_date?.split("T")[0] || "-"}</td>
                  <td><span className={`status status-${a.is_active ? "present" : "absent"}`}>{a.is_active ? "Active" : "Inactive"}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
}
