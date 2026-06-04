import { useAuth } from "../context/AuthContext";

/** Render children only when the user has the required permission(s). */
export default function PermGate({ perm, any, children, fallback = null }) {
  const { hasPerm } = useAuth();
  const allowed = perm ? hasPerm(perm) : (any?.some((p) => hasPerm(p)) ?? true);
  return allowed ? children : fallback;
}
