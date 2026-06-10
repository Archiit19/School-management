import { createContext, useContext, useState, useCallback, useEffect, useMemo } from "react";
import { authApi } from "../api/client";
import { parseTokenClaims } from "../utils/jwt";

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
      setUser(null);
      setPermissions([]);
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    authApi
      .me()
      .then((data) => {
        if (cancelled) return;
        const { token: refreshedToken, ...profile } = data;
        setUser(profile);
        setPermissions(profile.permissions || []);
        if (refreshedToken && refreshedToken !== token) {
          localStorage.setItem("token", refreshedToken);
          setToken(refreshedToken);
        }
      })
      .catch(() => {
        if (!cancelled) logout();
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [token, logout]);

  const claims = useMemo(() => parseTokenClaims(token), [token]);
  const sessionRole = user?.role_name || claims.role_name;
  const sessionSchoolId = user?.school_id || claims.school_id;

  const inSchoolContext = sessionRole === "super_admin" && Boolean(sessionSchoolId);
  const isPlatformAdmin = sessionRole === "platform_admin" && !sessionSchoolId;

  const hasPerm = useCallback(
    (perm) => {
      if (user?.role_name === "super_admin") return true;
      return permissions.includes(perm);
    },
    [user, permissions],
  );

  const value = useMemo(
    () => ({
      token,
      user,
      permissions,
      loading,
      saveToken,
      logout,
      setUser,
      hasPerm,
      isPlatformAdmin,
      inSchoolContext,
    }),
    [token, user, permissions, loading, saveToken, logout, hasPerm, isPlatformAdmin, inSchoolContext],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside AuthProvider");
  return ctx;
}
