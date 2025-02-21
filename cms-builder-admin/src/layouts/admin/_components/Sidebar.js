import React from "react";
import { Link } from "react-router-dom";

import {
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
} from "@mui/material";

import DashboardIcon from "@mui/icons-material/Dashboard";
import FolderIcon from "@mui/icons-material/Folder";
import RestoreIcon from "@mui/icons-material/Restore";

const routes = [
  {
    name: "Models",
    path: "/admin/models",
    icon: <DashboardIcon />,
  },
  {
    name: "Media",
    path: "/admin/media",
    icon: <FolderIcon />,
  },
  {
    name: "Timeline",
    path: "/admin/timeline",
    icon: <RestoreIcon />,
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
