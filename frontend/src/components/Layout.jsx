import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { authApi } from "../api/client";

const PLATFORM_NAV = [
  { to: "/", icon: "grid", label: "Dashboard" },
  { to: "/schools", icon: "building", label: "Schools" },
];

const SCHOOL_NAV = [
  { to: "/", icon: "grid", label: "Dashboard" },
  { to: "/me", icon: "user-check", label: "My Portal", perms: ["view_own_profile", "view_own_attendance", "view_own_exams", "view_own_results", "view_own_assignments", "view_own_dues"] },
  { to: "/roles", icon: "shield", label: "Roles & Permissions", perms: ["create_role", "manage_permissions"] },
  { to: "/users", icon: "users", label: "Users & Students", perms: ["create_user", "view_users", "admit_student", "view_students"] },
  { to: "/academic", icon: "book-open", label: "Academic Structure", perms: ["create_class", "create_section", "create_subject", "view_academic"] },
  { to: "/teacher-assignments", icon: "user-check", label: "Teacher Assign", perms: ["assign_teacher", "view_academic"] },
  { to: "/attendance", icon: "calendar-check", label: "Attendance", perms: ["mark_attendance", "view_attendance", "mark_teacher_attendance", "view_teacher_attendance", "mark_own_teacher_attendance"] },
  { to: "/assignments", icon: "file-text", label: "Assignments", perms: ["create_assignment", "view_assignments", "submit_assignment"] },
  { to: "/exams", icon: "clipboard", label: "Exams & Results", perms: ["create_exam", "view_exams", "enter_marks", "publish_results", "view_results"] },
  { to: "/finance", icon: "dollar-sign", label: "Finance", perms: ["create_fee", "record_payment", "view_dues"] },
];

function Icon({ name }) {
  return <span className="nav-icon" data-icon={name} />;
}

export default function Layout() {
  const { user, logout, hasPerm, inSchoolContext, isPlatformAdmin, saveToken, loading } = useAuth();
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate("/login");
  }

  async function handleExitSchool() {
    try {
      const res = await authApi.exitSchool();
      saveToken(res.token);
      navigate("/schools");
    } catch {
      navigate("/schools");
    }
  }

  const showPlatformNav = !loading && isPlatformAdmin;
  const navSource = showPlatformNav ? PLATFORM_NAV : SCHOOL_NAV;

  const visibleNav = navSource.filter((item) => {
    if (item.to === "/me") {
      return user?.role_name === "student" || user?.role_name === "parent";
    }
    if (!item.perms) return true;
    return item.perms.some((p) => hasPerm(p));
  });

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <span className="brand-icon">S</span>
          <span className="brand-text">SchoolMgmt</span>
        </div>

        {inSchoolContext && user?.school && (
          <div className="text-sm" style={{ padding: "0 16px 12px", color: "var(--clr-muted)" }}>
            <strong style={{ color: "var(--clr-text)" }}>{user.school.name}</strong>
          </div>
        )}

        <nav className="sidebar-nav">
          {visibleNav.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/"}
              className={({ isActive }) => `nav-link ${isActive ? "active" : ""}`}
            >
              <Icon name={item.icon} />
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>

        <div className="sidebar-footer">
          {user && (
            <div className="user-badge">
              <div className="user-avatar">{user.name?.[0] || "U"}</div>
              <div className="user-info">
                <span className="user-name">{user.name}</span>
                <span className="user-role">{user.role_name}</span>
              </div>
            </div>
          )}
          {inSchoolContext && (
            <button className="btn btn-ghost btn-sm" style={{ marginBottom: 8 }} onClick={handleExitSchool}>
              Exit school
            </button>
          )}
          <button className="btn btn-ghost btn-sm" onClick={handleLogout}>
            Logout
          </button>
        </div>
      </aside>

      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
}
