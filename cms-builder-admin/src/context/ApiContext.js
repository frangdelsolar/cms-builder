import { useMemo, createContext } from "react";
import apiService from "../services/ApiService";
import { useAuth } from "./AuthContext";

const ApiContext = createContext({});

const ApiProvider = ({ children }) => {
  const { token } = useAuth();

  const service = useMemo(() => {
    return apiService({
      token,
      apiBaseUrl: process.env.REACT_APP_API_BASE_URL,
    });
  }, [token]);

  return <ApiContext.Provider value={service}>{children}</ApiContext.Provider>;
};

export { ApiContext, ApiProvider };
