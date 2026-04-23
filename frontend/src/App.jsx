import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider, useAuth } from "./context/AuthContext";
import Layout from "./components/Layout";
import LoginPage from "./pages/LoginPage";
import DashboardPage from "./pages/DashboardPage";
import UsersPage from "./pages/UsersPage";
import RolesPage from "./pages/RolesPage";
import AcademicPage from "./pages/AcademicPage";
import StudentsPage from "./pages/StudentsPage";
import TeacherAssignmentsPage from "./pages/TeacherAssignmentsPage";
import AttendancePage from "./pages/AttendancePage";
import AssignmentsPage from "./pages/AssignmentsPage";
import ExamsPage from "./pages/ExamsPage";
import FinancePage from "./pages/FinancePage";

function ProtectedRoute({ children }) {
  const { token, loading } = useAuth();
  if (loading) return <div style={{ padding: 40, textAlign: "center" }}>Loading...</div>;
  if (!token) return <Navigate to="/login" replace />;
  return children;
}

function GuestRoute({ children }) {
  const { token, loading } = useAuth();
  if (loading) return <div style={{ padding: 40, textAlign: "center" }}>Loading...</div>;
  if (token) return <Navigate to="/" replace />;
  return children;
}

function RequirePerm({ any, children }) {
  const { hasPerm, loading } = useAuth();
  if (loading) return null;
  if (any.length > 0 && !any.some((p) => hasPerm(p))) {
    return (
      <div style={{ padding: 40, textAlign: "center" }}>
        <h2>Access Denied</h2>
        <p>You don't have permission to view this page.</p>
      </div>
    );
  }
  return children;
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<GuestRoute><LoginPage /></GuestRoute>} />
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route index element={<DashboardPage />} />
            <Route path="users" element={<RequirePerm any={["create_user", "view_users"]}><UsersPage /></RequirePerm>} />
            <Route path="roles" element={<RequirePerm any={["create_role", "manage_permissions"]}><RolesPage /></RequirePerm>} />
            <Route path="academic" element={<RequirePerm any={["create_class", "create_section", "create_subject", "view_academic"]}><AcademicPage /></RequirePerm>} />
            <Route path="students" element={<RequirePerm any={["admit_student", "view_students"]}><StudentsPage /></RequirePerm>} />
            <Route path="teacher-assignments" element={<RequirePerm any={["assign_teacher"]}><TeacherAssignmentsPage /></RequirePerm>} />
            <Route path="attendance" element={<RequirePerm any={["mark_attendance", "view_attendance"]}><AttendancePage /></RequirePerm>} />
            <Route path="assignments" element={<RequirePerm any={["create_assignment", "view_assignments", "submit_assignment"]}><AssignmentsPage /></RequirePerm>} />
            <Route path="exams" element={<RequirePerm any={["create_exam", "enter_marks", "publish_results", "view_results"]}><ExamsPage /></RequirePerm>} />
            <Route path="finance" element={<RequirePerm any={["create_fee", "record_payment", "view_dues"]}><FinancePage /></RequirePerm>} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
