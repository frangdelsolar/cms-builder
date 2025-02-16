import { useMemo, createContext } from "react";
import apiService from "../services/ApiService";
import { useAuth } from "./AuthContext";
import { useAppSelector } from "../store/Hooks";
import { selectProjectData } from "../store/ProjectSlice";

const ApiContext = createContext({});

const ApiProvider = ({ children }) => {
  const { token } = useAuth();

  const projectData = useAppSelector(selectProjectData);

  const service = useMemo(() => {
    return apiService({
      token,
      apiBaseUrl: projectData.apiBaseUrl || "http://localhost:80",
    });
  }, [token]);

  return <ApiContext.Provider value={service}>{children}</ApiContext.Provider>;
};

export { ApiContext, ApiProvider };
