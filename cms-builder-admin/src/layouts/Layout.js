import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";

import Footer from "./_components/Footer";
import Header from "./_components/Header";
import Sidebar from "./_components/Sidebar";

import Drawer from "@mui/material/Drawer";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import CssBaseline from "@mui/material/CssBaseline";

import { useState } from "react";

import { Outlet } from "react-router-dom";

import { styled } from "@mui/material/styles";

const StyledContainer = styled(Container)(({ theme }) => ({
  padding: 8,
  maxWidth: "100% !important",
  // minHeight: "100vh",
  width: "100vw",
  backgroundColor: theme.palette.background.default,
  display: "flex",
  flexDirection: "column",
}));

export default function MainLayout() {
  const drawerWidth = 240;
  const toolbarHeight = 64;

  const [openDrawer, setOpenDrawer] = useState(false);

  const onMenuClick = () => {
    setOpenDrawer(!openDrawer);
  };

  const closeDrawer = () => {
    setOpenDrawer(false);
  };

  return (
    <Box sx={{ display: "flex", flexDirection: "column", minHeight: "100vh" }}>
      <CssBaseline />
      <Header handleDrawerToggle={onMenuClick} />
      <Drawer
        open={openDrawer}
        onClose={closeDrawer}
        sx={{
          width: drawerWidth,
          flexShrink: 0,
          [`& .MuiDrawer-paper`]: {
            width: drawerWidth,
            boxSizing: "border-box",
          },
        }}
      >
        <Box sx={{ overflow: "auto" }}>
          <Box sx={{ height: toolbarHeight }} />
          <Sidebar close={closeDrawer} />
        </Box>
      </Drawer>
      <Box
        sx={{
          display: "flex",
          flexDirection: "column",
          flexGrow: 1,
          minHeight: 0,
          marginLeft: openDrawer ? `${drawerWidth}px` : 0,
          transition: "0.2s",
        }}
        component="main"
      >
        <StyledContainer
          sx={{
            flexGrow: 1,
            overflowY: "auto",
          }}
        >
          <Outlet />
        </StyledContainer>
        <Footer />
      </Box>
    </Box>
  );
}
