import { useState, useEffect, useCallback } from "react";
import { financeApi, academicApi } from "../api/client";

export default function FinancePage() {
  const [tab, setTab] = useState("dues");
  const [dues, setDues] = useState([]);
  const [classes, setClasses] = useState([]);
  const [query, setQuery] = useState({ student_id: "", class_id: "", section_id: "" });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [busy, setBusy] = useState(false);

  const [feeForm, setFeeForm] = useState({ title: "", description: "", amount: "", class_id: "", section_id: "", student_id: "", due_date: "" });
  const [payForm, setPayForm] = useState({ fee_id: "", student_id: "", amount_paid: "", payment_date: "", method: "", reference: "" });

  const load = useCallback(async () => {
    try { setDues(await financeApi.getDues(query)); }
    catch (err) { setError(err.message); }
  }, [query]);

  useEffect(() => { load(); }, [load]);
  useEffect(() => { academicApi.getClasses().then(setClasses).catch(() => {}); }, []);

  const flatClasses = classes.map((c) => c.class || c);

  function msg(txt) { setSuccess(txt); setError(""); setTimeout(() => setSuccess(""), 3000); }

  async function handleCreateFee(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try {
      const payload = { ...feeForm, amount: parseFloat(feeForm.amount) };
      ["class_id", "section_id", "student_id", "due_date"].forEach((k) => { if (!payload[k]) delete payload[k]; });
      await financeApi.createFee(payload);
      msg("Fee created.");
      setFeeForm({ title: "", description: "", amount: "", class_id: "", section_id: "", student_id: "", due_date: "" });
      load();
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  async function handlePayment(e) {
    e.preventDefault(); setError(""); setBusy(true);
    try {
      const payload = { ...payForm, amount_paid: parseFloat(payForm.amount_paid) };
      await financeApi.recordPayment(payload);
      msg("Payment recorded.");
      setPayForm({ fee_id: "", student_id: "", amount_paid: "", payment_date: "", method: "", reference: "" });
      load();
    } catch (err) { setError(err.message); } finally { setBusy(false); }
  }

  return (
    <>
      <div className="page-header">
        <h1>Finance</h1>
        <p>Flow 9 — Manage fees, record payments, and view dues.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}
      {success && <div className="alert alert-success">{success}</div>}

      <div className="tabs">
        <button className={`tab ${tab === "dues" ? "active" : ""}`} onClick={() => setTab("dues")}>Dues</button>
        <button className={`tab ${tab === "fee" ? "active" : ""}`} onClick={() => setTab("fee")}>Create Fee</button>
        <button className={`tab ${tab === "payment" ? "active" : ""}`} onClick={() => setTab("payment")}>Record Payment</button>
      </div>

      {tab === "dues" && (
        <div className="card">
          <div className="card-title">Outstanding Dues <span className="badge badge-get">GET /dues</span></div>
          <div className="grid-3 mb-4">
            <div className="form-group"><label>Student ID</label><input placeholder="UUID..." value={query.student_id} onChange={(e) => setQuery((q) => ({ ...q, student_id: e.target.value }))} /></div>
            <div className="form-group">
              <label>Class</label>
              <select value={query.class_id} onChange={(e) => setQuery((q) => ({ ...q, class_id: e.target.value }))}>
                <option value="">All</option>
                {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
            </div>
            <div className="form-group"><label>Section ID</label><input placeholder="UUID..." value={query.section_id} onChange={(e) => setQuery((q) => ({ ...q, section_id: e.target.value }))} /></div>
          </div>
          <div className="table-wrap">
            <table>
              <thead><tr><th>Fee</th><th>Amount</th><th>Paid</th><th>Balance</th><th>Status</th><th>Due Date</th><th>Fee ID</th></tr></thead>
              <tbody>
                {(!dues || dues.length === 0) && <tr><td colSpan={7} className="empty">No dues found.</td></tr>}
                {(dues || []).map((d, i) => (
                  <tr key={i}>
                    <td><strong>{d.title}</strong></td>
                    <td>{d.amount?.toFixed(2)}</td>
                    <td>{d.paid_amount?.toFixed(2)}</td>
                    <td style={{ fontWeight: 600 }}>{d.balance?.toFixed(2)}</td>
                    <td>
                      <span className={`status ${d.status === "paid" ? "status-paid" : d.status === "partial" ? "status-partial" : "status-absent"}`}>
                        {d.status}
                      </span>
                    </td>
                    <td>{d.due_date ? d.due_date.split("T")[0] : "—"}</td>
                    <td><span className="mono truncate">{d.fee_id}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {tab === "fee" && (
        <div className="card">
          <div className="card-title">Create Fee <span className="badge badge-post">POST /fees</span></div>
          <form onSubmit={handleCreateFee}>
            <div className="grid-3">
              <div className="form-group"><label>Title</label><input required value={feeForm.title} onChange={(e) => setFeeForm((p) => ({ ...p, title: e.target.value }))} placeholder="Tuition Q1" /></div>
              <div className="form-group"><label>Amount</label><input type="number" required min="0.01" step="0.01" value={feeForm.amount} onChange={(e) => setFeeForm((p) => ({ ...p, amount: e.target.value }))} placeholder="5000" /></div>
              <div className="form-group"><label>Due Date (opt)</label><input type="date" value={feeForm.due_date} onChange={(e) => setFeeForm((p) => ({ ...p, due_date: e.target.value }))} /></div>
            </div>
            <div className="grid-4">
              <div className="form-group"><label>Description</label><input value={feeForm.description} onChange={(e) => setFeeForm((p) => ({ ...p, description: e.target.value }))} placeholder="Details" /></div>
              <div className="form-group">
                <label>Class (opt)</label>
                <select value={feeForm.class_id} onChange={(e) => setFeeForm((p) => ({ ...p, class_id: e.target.value }))}>
                  <option value="">Any</option>
                  {flatClasses.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
              </div>
              <div className="form-group"><label>Section ID (opt)</label><input value={feeForm.section_id} onChange={(e) => setFeeForm((p) => ({ ...p, section_id: e.target.value }))} placeholder="UUID" /></div>
              <div className="form-group"><label>Student ID (opt)</label><input value={feeForm.student_id} onChange={(e) => setFeeForm((p) => ({ ...p, student_id: e.target.value }))} placeholder="UUID" /></div>
            </div>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Create Fee</button></div>
          </form>
        </div>
      )}

      {tab === "payment" && (
        <div className="card">
          <div className="card-title">Record Payment <span className="badge badge-post">POST /payments</span></div>
          <form onSubmit={handlePayment}>
            <div className="grid-3">
              <div className="form-group"><label>Fee ID</label><input required value={payForm.fee_id} onChange={(e) => setPayForm((p) => ({ ...p, fee_id: e.target.value }))} placeholder="UUID" /></div>
              <div className="form-group"><label>Student ID</label><input required value={payForm.student_id} onChange={(e) => setPayForm((p) => ({ ...p, student_id: e.target.value }))} placeholder="UUID" /></div>
              <div className="form-group"><label>Amount Paid</label><input type="number" required min="0.01" step="0.01" value={payForm.amount_paid} onChange={(e) => setPayForm((p) => ({ ...p, amount_paid: e.target.value }))} placeholder="5000" /></div>
            </div>
            <div className="grid-3">
              <div className="form-group"><label>Payment Date</label><input type="date" required value={payForm.payment_date} onChange={(e) => setPayForm((p) => ({ ...p, payment_date: e.target.value }))} /></div>
              <div className="form-group"><label>Method</label><input value={payForm.method} onChange={(e) => setPayForm((p) => ({ ...p, method: e.target.value }))} placeholder="cash / card / UPI" /></div>
              <div className="form-group"><label>Reference</label><input value={payForm.reference} onChange={(e) => setPayForm((p) => ({ ...p, reference: e.target.value }))} placeholder="TXN-001" /></div>
            </div>
            <div className="btn-row"><button className="btn btn-primary" disabled={busy}>Record Payment</button></div>
          </form>
        </div>
      )}
    </>
  );
}
