import { BrowserRouter } from "react-router-dom";
import { ThemeProvider } from "@mui/material/styles";

import AppRoutes from "./Routes";
import theme from "./context/ThemeContext";
import StoreProvider from "./store/StoreProvider";
import { ApiProvider } from "./context/ApiContext";
import { AuthProvider } from "./context/AuthContext";
import { DialogProvider } from "./context/DialogContext";
import { ToastProvider } from "./context/ToastContext";

export default function App() {
  return (
    <StoreProvider>
      <AuthProvider>
        <ApiProvider>
          <ThemeProvider theme={theme}>
            <BrowserRouter>
              <ToastProvider>
                <DialogProvider>
                  <AppRoutes />
                </DialogProvider>
              </ToastProvider>
            </BrowserRouter>
          </ThemeProvider>
        </ApiProvider>
      </AuthProvider>
    </StoreProvider>
  );
}
