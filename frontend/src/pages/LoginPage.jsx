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
  const [signupForm, setSignupForm] = useState({ name: "", email: "", password: "" });

  function field(setter) {
    return (e) => setter((p) => ({ ...p, [e.target.name]: e.target.value }));
  }

  function homeFor(user) {
    return user?.role_name === "platform_admin" ? "/schools" : "/";
  }

  async function handleLogin(e) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      const res = await authApi.login(loginForm);
      saveToken(res.token);
      navigate(homeFor(res.user));
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  async function handleSignup(e) {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      const res = await authApi.signup(signupForm);
      saveToken(res.token);
      navigate(homeFor(res.user));
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
          <p>Sign in or create an admin account, then add your schools from the dashboard.</p>

          <div className="login-tabs">
            <button className={`login-tab ${tab === "login" ? "active" : ""}`} onClick={() => { setTab("login"); setError(""); }}>
              Login
            </button>
            <button className={`login-tab ${tab === "signup" ? "active" : ""}`} onClick={() => { setTab("signup"); setError(""); }}>
              Sign Up
            </button>
          </div>

          {error && <div className="alert alert-error">{error}</div>}

          {tab === "login" ? (
            <form onSubmit={handleLogin}>
              <div className="form-group">
                <label>Email</label>
                <input name="email" type="email" required value={loginForm.email} onChange={field(setLoginForm)} placeholder="you@example.com" />
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
            <form onSubmit={handleSignup}>
              <div className="form-group">
                <label>Full Name</label>
                <input name="name" required value={signupForm.name} onChange={field(setSignupForm)} placeholder="John Doe" />
              </div>
              <div className="form-group">
                <label>Email</label>
                <input name="email" type="email" required value={signupForm.email} onChange={field(setSignupForm)} placeholder="you@example.com" />
              </div>
              <div className="form-group">
                <label>Password</label>
                <input name="password" type="password" required minLength={6} value={signupForm.password} onChange={field(setSignupForm)} placeholder="Min 6 characters" />
              </div>
              <button className="btn btn-primary" disabled={busy}>
                {busy ? "Creating account..." : "Create Admin Account"}
              </button>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
