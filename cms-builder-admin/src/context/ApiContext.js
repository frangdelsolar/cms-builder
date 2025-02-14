import { useMemo, createContext } from "react";
import apiService from "../services/ApiService";
import { useAuth } from "../context/AuthContext";

const ApiContext = createContext({
  getEntities: () => {},
  getFiles: () => {},
  deleteFile: () => {},
  downloadFile: () => {},
  getFileInfo: () => {},
  getEndpoints: () => {},
  schema: () => {},
  post: () => {},
  put: () => {},
  destroy: () => {},
  list: () => {},
});

const ApiProvider = ({ children }) => {
  const { token } = useAuth();

  const service = useMemo(() => {
    return apiService({ token });
  }, [token]);

  return <ApiContext.Provider value={service}>{children}</ApiContext.Provider>;
};

export { ApiContext, ApiProvider };
