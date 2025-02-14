import React, { createContext, useState, useContext, useEffect } from "react";
import firebaseLogin from "../services/FirebaseService";

const AuthContext = createContext({
  isAuthenticated: false,
  user: null,
  token: "",
  login: () => {},
  logout: () => {},
});

export const timeLeft = (user) => {
  if (user && user.expiresIn && user.storedAt) {
    const expiryTime = user.storedAt + parseInt(user.expiresIn, 10) * 1000;
    return expiryTime - Date.now();
  }
  return 0; // If no expiry info, return 0
};

const isTokenValid = (user) => {
  return timeLeft(user) > 0;
};

export const AuthProvider = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState(null);
  const [token, setToken] = useState("");
  const [isLoading, setIsLoading] = useState(true);

  const [isRefreshing, setIsRefreshing] = useState(false); // Add refreshing state

  const refreshAuthStatus = async () => {
    if (isRefreshing) return;

    setIsRefreshing(true);

    try {
      const storedUser = localStorage.getItem("user");
      const storedToken = localStorage.getItem("token");

      if (storedUser && storedToken) {
        const parsedUser = JSON.parse(storedUser);

        if (isTokenValid(parsedUser)) {
          // TODO: Refresh token logic (if Firebase supports it)
        } else {
          logout();
        }
      }
    } catch (error) {
      console.error("Error refreshing auth:", error);
      logout();
    } finally {
      setIsRefreshing(false);
    }
  };

  const startRefreshInterval = (user) => {
    // Separate function for interval setup
    const remainingTime = timeLeft(user);
    console.log(
      "Auth token expiring in about",
      Math.floor(remainingTime / 1000 / 60),
      "minutes"
    );

    if (isNaN(remainingTime)) {
      console.warn("Invalid timeLeft value. Using fallback refresh interval.");
      return setInterval(refreshAuthStatus, 5 * 60 * 1000); // 5 minutes fallback
    }

    const refreshBefore = remainingTime - 4 * 60 * 1000;
    return setInterval(
      refreshAuthStatus,
      refreshBefore > 0 ? refreshBefore : 5 * 60 * 1000
    );
  };

  useEffect(() => {
    const checkAuthStatus = async () => {
      const storedAuth = localStorage.getItem("isAuthenticated");
      const storedUser = localStorage.getItem("user");
      const storedToken = localStorage.getItem("token");

      if (storedAuth === "true" && storedUser && storedToken) {
        try {
          // Add try catch block for the JSON.parse
          const parsedUser = JSON.parse(storedUser);

          if (isTokenValid(parsedUser)) {
            setIsAuthenticated(true);
            setUser(parsedUser);
            setToken(storedToken);

            const intervalId = startRefreshInterval(parsedUser); // Call the separate function
            return () => clearInterval(intervalId);
          } else {
            logout();
          }
        } catch (error) {
          console.error("Error parsing stored user data:", error);
          logout();
        }
      }

      setIsAuthenticated(false);
      setUser(null);
      setToken("");
    };

    checkAuthStatus();
    setIsLoading(false);
  }, []);

  const login = async (userData) => {
    try {
      const response = await firebaseLogin(userData.email, userData.password);

      if (response) {
        setIsAuthenticated(true);
        setUser(response.user);
        setToken(response.idToken);

        localStorage.setItem("isAuthenticated", "true");
        localStorage.setItem("user", JSON.stringify(response.user)); // Store updated user object
        localStorage.setItem("token", response.idToken);
      } else {
        console.error("Login failed");
      }
    } catch (error) {
      console.error("Login error:", error);
    }
  };

  const logout = () => {
    setIsAuthenticated(false);
    setUser(null);
    setToken("");
    localStorage.removeItem("isAuthenticated");
    localStorage.removeItem("user");
    localStorage.removeItem("token");
  };

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        user,
        token,
        isLoading,
        login,
        logout,
        refreshAuthStatus,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
