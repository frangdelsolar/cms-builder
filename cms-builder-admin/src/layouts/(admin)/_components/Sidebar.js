import React from "react";
import { Link } from "react-router-dom";

import {
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
} from "@mui/material";

import HomeIcon from "@mui/icons-material/Home";
import DashboardIcon from "@mui/icons-material/Dashboard";
import FolderIcon from "@mui/icons-material/Folder";

const routes = [
  {
    name: "Home",
    path: "/",
    icon: <HomeIcon />,
  },
  {
    name: "Models",
    path: "/models",
    icon: <DashboardIcon />,
  },
  {
    name: "Media",
    path: "/media",
    icon: <FolderIcon />,
  },
];

function Sidebar(props) {
  return (
    <List>
      {routes.map((route) => (
        <ListItem key={route.name} disablePadding>
          <ListItemButton
            component={Link}
            to={route.path}
            onClick={props.close}
          >
            <ListItemIcon>{route.icon}</ListItemIcon>
            <ListItemText primary={route.name} />
          </ListItemButton>
        </ListItem>
      ))}
    </List>
  );
}

export default Sidebar;
