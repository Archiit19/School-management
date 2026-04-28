import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

const NAV = [
  { to: "/", icon: "grid", label: "Dashboard", perms: [] },
  { to: "/me", icon: "user-check", label: "My Portal", perms: ["view_own_profile", "view_own_attendance", "view_own_results", "view_own_assignments", "view_own_dues"] },
  { to: "/users", icon: "users", label: "Users", perms: ["create_user", "view_users"] },
  { to: "/roles", icon: "shield", label: "Roles & Permissions", perms: ["create_role", "manage_permissions"] },
  { to: "/academic", icon: "book-open", label: "Academic Structure", perms: ["create_class", "create_section", "create_subject", "view_academic"] },
  { to: "/students", icon: "graduation-cap", label: "Students", perms: ["admit_student", "view_students"] },
  { to: "/teacher-assignments", icon: "user-check", label: "Teacher Assign", perms: ["assign_teacher"] },
  { to: "/attendance", icon: "calendar-check", label: "Attendance", perms: ["mark_attendance", "view_attendance", "mark_teacher_attendance", "view_teacher_attendance", "mark_own_teacher_attendance"] },
  { to: "/assignments", icon: "file-text", label: "Assignments", perms: ["create_assignment", "view_assignments", "submit_assignment"] },
  { to: "/exams", icon: "clipboard", label: "Exams & Results", perms: ["create_exam", "enter_marks", "publish_results", "view_results"] },
  { to: "/finance", icon: "dollar-sign", label: "Finance", perms: ["create_fee", "record_payment", "view_dues"] },
];

function Icon({ name }) {
  return <span className="nav-icon" data-icon={name} />;
}

export default function Layout() {
  const { user, logout, hasPerm } = useAuth();
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate("/login");
  }

  const visibleNav = NAV.filter(
    (item) => item.perms.length === 0 || item.perms.some((p) => hasPerm(p)),
  );

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <span className="brand-icon">S</span>
          <span className="brand-text">SchoolMgmt</span>
        </div>

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
