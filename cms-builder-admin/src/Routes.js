import MainLayout from "./layouts/admin/Layout";
import AuthLayout from "./layouts/auth/Layout";

import LoginPage from "./layouts/auth/Login/Page";
import RegisterPage from "./layouts/auth/Register/Page";
import ForgotPasswordPage from "./layouts/auth/ForgotPassword/Page";

import ModelPage from "./layouts/admin/Models/Page";
import MediaPage from "./layouts/admin/Media/Page";

import { useRoutes, Navigate } from "react-router-dom";
import { useAuth } from "./context/AuthContext";
import TimelinePage from "./layouts/admin/Timeline/Page";

const PrivateRoute = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    // TODO: Show a loading spinner or something
    return <div>Loading...</div>;
  }

  return isAuthenticated && !isLoading ? (
    children
  ) : (
    <Navigate to="/auth/login" />
  );
};

const RedirectHome = () => {
  return <Navigate to="admin/models" />;
};

export const ROUTES = [
  {
    path: "/",
    element: <RedirectHome />,
  },

  {
    path: "admin",
    element: (
      <PrivateRoute>
        <MainLayout />
      </PrivateRoute>
    ),
    children: [
      {
        path: "models",
        element: <ModelPage />,
      },
      {
        path: "timeline",
        element: <TimelinePage />,
      },
      {
        path: "media",
        element: <MediaPage />,
      },
    ],
  },
  {
    path: "auth",
    element: <AuthLayout />,
    children: [
      {
        path: "login",
        element: <LoginPage />,
      },
      {
        path: "register",
        element: <RegisterPage />,
      },
      {
        path: "forgot-password",
        element: <ForgotPasswordPage />,
      },
    ],
  },
];

function AppRoutes() {
  const elements = useRoutes(ROUTES);
  return <>{elements}</>;
}

export default AppRoutes;
