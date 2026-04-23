import { createContext, useContext, useState, useCallback, useEffect } from "react";
import { authApi } from "../api/client";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() => localStorage.getItem("token") || "");
  const [user, setUser] = useState(null);
  const [permissions, setPermissions] = useState([]);
  const [loading, setLoading] = useState(!!localStorage.getItem("token"));

  const saveToken = useCallback((t) => {
    localStorage.setItem("token", t);
    setToken(t);
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem("token");
    setToken("");
    setUser(null);
    setPermissions([]);
  }, []);

  useEffect(() => {
    if (!token) {
      setLoading(false);
      return;
    }
    authApi
      .me()
      .then((data) => {
        setUser(data);
        setPermissions(data.permissions || []);
      })
      .catch(() => logout())
      .finally(() => setLoading(false));
  }, [token, logout]);

  const hasPerm = useCallback(
    (perm) => {
      if (user?.role_name === "super_admin") return true;
      return permissions.includes(perm);
    },
    [user, permissions],
  );

  const value = { token, user, permissions, loading, saveToken, logout, setUser, hasPerm };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
  return ctx;
}
