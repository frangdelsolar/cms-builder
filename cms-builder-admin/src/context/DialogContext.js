import React, { createContext, useContext, useState } from "react";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Paper,
} from "@mui/material";

const DialogContext = createContext();

const useDialogs = () => {
  const context = useContext(DialogContext);
  if (!context) {
    throw new Error("useDialog must be used within a DialogProvider");
  }
  return context;
};

const DialogProvider = ({ children }) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [dialogTitle, setDialogTitle] = useState("");
  const [dialogContent, setDialogContent] = useState(null); // Can be JSX or string
  const [dialogActions, setDialogActions] = useState(null); // Array of action objects

  const show = ({ title, content, actions }) => {
    setDialogTitle(title || "");
    setDialogContent(content || null);
    setDialogActions(actions || null);
    setDialogOpen(true);
  };

  const close = () => {
    setDialogOpen(false);
  };

  const confirm = ({ title, content }) => {
    return new Promise((resolve) => {
      show({
        title: title || "Confirmar", // Default title
        content: content || "¿Está seguro?", // Default content
        actions: [
          {
            label: "Cancelar",
            onClick: () => {
              close();
              resolve(false); // Resolve with false if cancelled
            },
          },
          {
            label: "OK",
            onClick: () => {
              close();
              resolve(true); // Resolve with true if confirmed
            },
            color: "primary",
          },
        ],
      });
    });
  };

  return (
    <DialogContext.Provider value={{ show, close, confirm }}>
      {children}
      <Dialog open={dialogOpen} onClose={close} maxWidth="sm" fullWidth>
        {" "}
        <Paper elevation={0} sx={{ width: "100%" }}>
          {" "}
          <DialogTitle>{dialogTitle}</DialogTitle>
          <DialogContent>
            {
              typeof dialogContent === "string" ? (
                <>{dialogContent}</>
              ) : (
                dialogContent
              ) // Or render something else for other types
            }
          </DialogContent>
          {dialogActions && (
            <DialogActions>
              {dialogActions.map((action, index) => (
                <Button
                  key={index}
                  onClick={action.onClick}
                  color={action.color || "primary"}
                >
                  {action.label}
                </Button>
              ))}
            </DialogActions>
          )}
        </Paper>
      </Dialog>
    </DialogContext.Provider>
  );
};

export { DialogProvider, useDialogs };
