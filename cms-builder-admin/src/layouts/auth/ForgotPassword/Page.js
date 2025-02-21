import React, { useState } from "react";
import {
  TextField,
  Button,
  Box,
  Typography,
  Alert,
  Paper,
} from "@mui/material";
import { useNavigate, Link } from "react-router-dom";

function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [message, setMessage] = useState(null);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    setMessage(null);
    // TODO: implement forgot password
    try {
      const response = await new Promise((resolve) =>
        setTimeout(() => resolve({ success: true }), 500)
      );

      if (response.success) {
        setMessage("A reset link has been sent to your email address.");
        navigate("/auth/login");
      } else {
        setError("Email not found. Please check your email address.");
      }
    } catch (err) {
      setError(err.message || "An error occurred. Please try again later.");
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
            Reset Password
          </Typography>

          {error && <Alert severity="error">{error}</Alert>}
          {message && <Alert severity="success">{message}</Alert>}

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

            <Button
              type="submit"
              variant="contained"
              color="primary"
              fullWidth
              sx={{ mt: 2 }}
            >
              Send Reset Link
            </Button>
          </form>

          <Typography variant="body2" sx={{ mt: 2 }}>
            Remember your password? <Link to="/auth/login">Login</Link>
          </Typography>
        </Box>
      </Paper>
    </Box>
  );
}

export default ForgotPasswordPage;
