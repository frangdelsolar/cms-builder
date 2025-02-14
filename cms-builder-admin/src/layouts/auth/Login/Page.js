import { useState } from "react";
import {
  TextField,
  Button,
  Box,
  Typography,
  Alert,
  Paper,
} from "@mui/material";
import { useAuth } from "../../../context/AuthContext";
import { useNavigate, Link } from "react-router-dom";

function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState(null);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);

    try {
      const userData = { email, password };
      await login(userData);
      navigate("/");
    } catch (err) {
      setError(err.message || "Invalid credentials");
      console.error(err);
    }
  };

  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        mt: 8,
        minHeight: "100%",
        width: "100%",
      }}
    >
      <Paper elevation={0} sx={{ padding: 4, width: "100%", maxWidth: 400 }}>
        <Box
          sx={{
            width: "100%",
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
            justifyContent: "center",
            minHeight: "100%",
          }}
        >
          <Typography variant="h5" component="h2" gutterBottom>
            Login
          </Typography>
          {error && <Alert severity="error">{error}</Alert>}
          <form onSubmit={handleSubmit} style={{ width: "100%" }}>
            <TextField
              label="Email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              fullWidth
              margin="normal"
              required
            />
            <TextField
              label="Password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              fullWidth
              margin="normal"
              required
            />

            <Button
              type="submit"
              variant="contained"
              color="primary"
              fullWidth
              sx={{ mt: 2 }}
            >
              Login
            </Button>
          </form>
          <Typography variant="body2" sx={{ mt: 2 }}>
            Don't have an account? <Link to="/auth/register">Register</Link>
          </Typography>
          <Typography variant="body2" sx={{ mt: 1 }}>
            Forgot your password?{" "}
            <Link to="/auth/forgot-password">Reset it</Link>
          </Typography>
        </Box>
      </Paper>
    </Box>
  );
}

export default LoginPage;
