import { useMemo, createContext, useState, useEffect } from "react";
import apiService from "../services/ApiService";
import { useAuth } from "./AuthContext";
import { useNotifications } from "./ToastContext";

// Create the API context
const ApiContext = createContext({});

const ApiProvider = ({ children }) => {
  const { token, user } = useAuth();
  const toast = useNotifications();

  const [roles, setRoles] = useState([]);
  const [healthy, setHealthy] = useState(false);
  const [svc, setSvc] = useState(null);

  // Initialize the API service
  const service = useMemo(() => {
    const svc = apiService({
      token,
      apiBaseUrl: process.env.REACT_APP_API_BASE_URL,
    });
    setSvc(svc);
    return svc;
  }, [token, healthy]);

  // Health check function
  const checkHealth = async () => {
    if (!svc) return;

    try {
      const res = await svc.healthCheck();
      if (res.success) {
        setHealthy(true);
      }
    } catch (error) {
      toast.show("Server may not be ready", "error");
      setHealthy(false);
    }
  };

  // Run health check on mount and when the service changes
  useEffect(() => {
    checkHealth();
  }, [svc]);

  useEffect(() => {
    if (!healthy) {
      return;
    }

    getUserRoles();
  }, [healthy]);

  // Fetch user roles
  const getUserRoles = async () => {
    if (!svc) return;

    const firebaseId = user?.uid;
    try {
      const response = await svc.list("users", 1, 1, "", {
        fbId: firebaseId,
      });
      let rs = response.data[0]?.roles.split(",") || [];

      setRoles(rs);
    } catch (error) {
      toast.show("Failed to fetch user roles", "error");
    }
  };

  // Provide the API service and any additional methods to the context
  const contextValue = {
    ...service,
    healthy,
    roles,
    getUserRoles,
  };

  return (
    <ApiContext.Provider value={contextValue}>{children}</ApiContext.Provider>
  );
};

export { ApiContext, ApiProvider };
