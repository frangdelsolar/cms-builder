import Box from "@mui/material/Box";
import { Outlet } from "react-router-dom";
import Container from "@mui/material/Container";
import { styled } from "@mui/material/styles";

const StyledContainer = styled(Container)(({ theme }) => ({
  padding: 0,
  maxWidth: "100% !important",
  minHeight: "100vh",
  width: "100vw",
  backgroundColor: theme.palette.background.default,
  display: "flex",
  flexDirection: "column",
}));

function AuthLayout() {
  return (
    <StyledContainer component="main" maxWidth="xs">
      <Box
        sx={{
          padding: 8,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
        }}
      >
        <Outlet />
      </Box>
    </StyledContainer>
  );
}

export default AuthLayout;
