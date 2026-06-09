import { Navigate } from "react-router-dom";

/** Students are now users with the student role — redirect to unified user management. */
export default function StudentsPage() {
  return <Navigate to="/users" replace />;
}
