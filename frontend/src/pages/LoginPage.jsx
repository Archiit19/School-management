import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { authApi } from "../api/client";
import { useAuth } from "../context/AuthContext";

export default function LoginPage() {
  const [tab, setTab] = useState("login");
  const { saveToken } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const [loginForm, setLoginForm] = useState({ email: "", password: "" });
  const [regForm, setRegForm] = useState({
    school_name: "",
    school_email: "",
    school_address: "",
    school_phone: "",
    admin_name: "",
    admin_email: "",
    admin_password: "",
  });

  function field(setter) {
    return (e) => setter((p) => ({ ...p, [e.target.name]: e.target.value }));
  }

  async function handleLogin(e) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      const res = await authApi.login(loginForm);
      saveToken(res.token);
      navigate("/");
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  async function handleRegister(e) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      const res = await authApi.registerSchool(regForm);
      saveToken(res.token);
      navigate("/");
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="login-page">
      <div className="login-container">
        <div className="login-card">
          <h1>School Management</h1>
          <p>Sign in to your account or register a new school.</p>

          <div className="login-tabs">
            <button className={`login-tab ${tab === "login" ? "active" : ""}`} onClick={() => { setTab("login"); setError(""); }}>
              Login
            </button>
            <button className={`login-tab ${tab === "register" ? "active" : ""}`} onClick={() => { setTab("register"); setError(""); }}>
              Register School
            </button>
          </div>

          {error && <div className="alert alert-error">{error}</div>}

          {tab === "login" ? (
            <form onSubmit={handleLogin}>
              <div className="form-group">
                <label>Email</label>
                <input name="email" type="email" required value={loginForm.email} onChange={field(setLoginForm)} placeholder="admin@school.edu" />
              </div>
              <div className="form-group">
                <label>Password</label>
                <input name="password" type="password" required value={loginForm.password} onChange={field(setLoginForm)} placeholder="Enter password" />
              </div>
              <button className="btn btn-primary" disabled={busy}>
                {busy ? "Signing in..." : "Sign In"}
              </button>
            </form>
          ) : (
            <form onSubmit={handleRegister}>
              <div className="form-group">
                <label>School Name</label>
                <input name="school_name" required value={regForm.school_name} onChange={field(setRegForm)} placeholder="Springfield Elementary" />
              </div>
              <div className="grid-2">
                <div className="form-group">
                  <label>School Email</label>
                  <input name="school_email" type="email" required value={regForm.school_email} onChange={field(setRegForm)} placeholder="info@school.edu" />
                </div>
                <div className="form-group">
                  <label>School Phone</label>
                  <input name="school_phone" value={regForm.school_phone} onChange={field(setRegForm)} placeholder="555-0100" />
                </div>
              </div>
              <div className="form-group">
                <label>School Address</label>
                <input name="school_address" value={regForm.school_address} onChange={field(setRegForm)} placeholder="123 Main St" />
              </div>
              <hr style={{ border: "none", borderTop: "1px solid var(--clr-border)", margin: "12px 0" }} />
              <div className="form-group">
                <label>Admin Full Name</label>
                <input name="admin_name" required value={regForm.admin_name} onChange={field(setRegForm)} placeholder="John Doe" />
              </div>
              <div className="grid-2">
                <div className="form-group">
                  <label>Admin Email</label>
                  <input name="admin_email" type="email" required value={regForm.admin_email} onChange={field(setRegForm)} placeholder="john@school.edu" />
                </div>
                <div className="form-group">
                  <label>Admin Password</label>
                  <input name="admin_password" type="password" required minLength={6} value={regForm.admin_password} onChange={field(setRegForm)} placeholder="Min 6 characters" />
                </div>
              </div>
              <button className="btn btn-primary" disabled={busy}>
                {busy ? "Registering..." : "Register School"}
              </button>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
