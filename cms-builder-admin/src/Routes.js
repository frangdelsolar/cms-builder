import MainLayout from "./layouts/admin/Layout";
import AuthLayout from "./layouts/auth/Layout";

import LoginPage from "./layouts/auth/Login/Page";
import RegisterPage from "./layouts/auth/Register/Page";
import ForgotPasswordPage from "./layouts/auth/ForgotPassword/Page";

import ModelPage from "./layouts/admin/Models/Page";
import MediaPage from "./layouts/admin/Media/Page";
import ActivityPage from "./layouts/admin/Activity/Page";

import { useState, useEffect, useContext } from "react";
import { useRoutes, Navigate } from "react-router-dom";
import { useAuth } from "./context/AuthContext";
import { ApiContext } from "./context/ApiContext";
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
  return <Navigate to="admin/activity" />;
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
        path: "activity",
        element: <ActivityPage />,
      },
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
  const [apiReady, setApiReady] = useState(false);
  const apiService = useContext(ApiContext);

  useEffect(() => {
    const healthCheck = async () => {
      try {
        const response = await apiService.healthCheck();
        if (response.ok) {
          console.log("API is healthy");
        } else {
          console.error("API is not healthy");
        }
      } catch (error) {
        console.error("Error checking API health:", error);
      }
    };

    if (!apiReady) {
      healthCheck().then(() => setApiReady(true));
    }
  }, [apiReady]);

  const elements = useRoutes(ROUTES);
  return <>{elements}</>;
}

export default AppRoutes;
