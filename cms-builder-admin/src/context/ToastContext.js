import React, { createContext, useContext, useEffect, useState } from "react";
import { Snackbar, Alert } from "@mui/material";

const ToastContext = createContext();

const useNotifications = () => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error(
      "useNotification must be used within a NotificationProvider"
    );
  }
  return context;
};

const ToastProvider = ({ children }) => {
  const [notificationsQueue, setNotificationsQueue] = useState([]); // The notification queue
  const [currentNotification, setCurrentNotification] = useState(null);

  const show = (message, severity = "info") => {
    const newNotification = { message, severity, key: new Date().getTime() };
    setNotificationsQueue((prevNotifications) => [
      ...prevNotifications,
      newNotification,
    ]);
  };

  useEffect(() => {
    if (notificationsQueue.length > 0 && !currentNotification) {
      const nextNotification = notificationsQueue[0];
      setCurrentNotification(nextNotification);
      setNotificationsQueue((prevQueue) => prevQueue.slice(1));
    }
  }, [notificationsQueue, currentNotification]);

  const handleClose = (key) => {
    if (currentNotification && currentNotification.key === key) {
      setCurrentNotification(null); // Clear current notification
    }
  };

  return (
    <ToastContext.Provider value={{ show }}>
      {children}
      {currentNotification && (
        <Snackbar
          key={currentNotification.key}
          open={true}
          autoHideDuration={6000}
          onClose={() => handleClose(currentNotification.key)}
          onExited={() => handleClose(currentNotification.key)}
          TransitionProps={{ appear: true }}
        >
          <Alert
            onClose={() => handleClose(currentNotification.key)}
            severity={currentNotification.severity}
            sx={{ width: "100%" }}
          >
            {currentNotification.message}
          </Alert>
        </Snackbar>
      )}
    </ToastContext.Provider>
  );
};

export { ToastProvider, useNotifications };
