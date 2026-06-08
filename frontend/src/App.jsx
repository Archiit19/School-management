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
import SchoolsPage from "./pages/SchoolsPage";
import MyPortalPage from "./pages/MyPortalPage";

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
  const { hasPerm, loading, user } = useAuth();
  if (loading || (user === null && any.length > 0)) return null;
  if (any.length > 0 && !any.some((p) => hasPerm(p))) {
    return <Navigate to="/" replace />;
  }
  return children;
}

function HomeRoute() {
  const { isPlatformAdmin, loading } = useAuth();
  if (loading) return <div style={{ padding: 40, textAlign: "center" }}>Loading...</div>;
  if (isPlatformAdmin) {
    return <Navigate to="/schools" replace />;
  }
  return <DashboardPage />;
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<GuestRoute><LoginPage /></GuestRoute>} />
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route index element={<HomeRoute />} />
            <Route path="users" element={<RequirePerm any={["create_user", "view_users"]}><UsersPage /></RequirePerm>} />
            <Route path="roles" element={<RequirePerm any={["create_role", "manage_permissions"]}><RolesPage /></RequirePerm>} />
            <Route path="academic" element={<RequirePerm any={["create_class", "create_section", "create_subject", "view_academic"]}><AcademicPage /></RequirePerm>} />
            <Route path="students" element={<RequirePerm any={["admit_student", "view_students"]}><StudentsPage /></RequirePerm>} />
            <Route path="teacher-assignments" element={<RequirePerm any={["assign_teacher"]}><TeacherAssignmentsPage /></RequirePerm>} />
            <Route path="attendance" element={<RequirePerm any={["mark_attendance", "view_attendance", "mark_teacher_attendance", "view_teacher_attendance", "mark_own_teacher_attendance"]}><AttendancePage /></RequirePerm>} />
            <Route path="assignments" element={<RequirePerm any={["create_assignment", "view_assignments", "submit_assignment"]}><AssignmentsPage /></RequirePerm>} />
            <Route path="exams" element={<RequirePerm any={["create_exam", "view_exams", "enter_marks", "publish_results", "view_results"]}><ExamsPage /></RequirePerm>} />
            <Route path="finance" element={<RequirePerm any={["create_fee", "record_payment", "view_dues"]}><FinancePage /></RequirePerm>} />
            <Route path="schools" element={<RequirePerm any={["view_my_schools", "create_school"]}><SchoolsPage /></RequirePerm>} />
            <Route
              path="me"
              element={
                <RequirePerm any={["view_own_profile", "view_own_attendance", "view_own_exams", "view_own_results", "view_own_assignments", "view_own_dues"]}>
                  <MyPortalPage />
                </RequirePerm>
              }
            />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
