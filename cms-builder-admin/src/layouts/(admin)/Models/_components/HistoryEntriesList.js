import { Card, CardHeader } from "@mui/material";
import { useEffect, useState, useContext } from "react";

import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";

import EditIcon from "@mui/icons-material/Edit";
import ClearIcon from "@mui/icons-material/Clear";
import AddIcon from "@mui/icons-material/Add";

import { ApiContext } from "../../../../context/ApiContext";

const HistoryEntriesList = () => {
  const apiService = useContext(ApiContext);

  const [entries, setEntries] = useState([]);

  useEffect(() => {
    let fn = async () => {
      try {
        const response = await apiService.list(
          "history-entries",
          1,
          10,
          "-timestamp"
        );
        setEntries(response.data);
      } catch (error) {
        console.log(error);
      }
    };

    fn();
  }, []);

  useEffect(() => {
    if (!entries) {
      return;
    }
  }, [entries]);

  const getIcon = (action) => {
    switch (action) {
      case "updated":
        return <EditIcon />;
      case "deleted":
        return <ClearIcon />;
      case "created":
        return <AddIcon />;
      default:
        return null;
    }
  };

  const generateListItem = (entry) => {
    let date = new Date(entry.timestamp);
    let resourceLabel = entry.resourceName + " (" + entry.resourceId + ")";
    return (
      <ListItem key={entry.ID}>
        <ListItemIcon>{getIcon(entry.action)}</ListItemIcon>
        <ListItemText
          primary={entry.username + " " + entry.action + " " + resourceLabel}
          secondary={
            date.toLocaleDateString() + " at " + date.toLocaleTimeString()
          }
        />
      </ListItem>
    );
  };

  return (
    <Card>
      <CardHeader title="Acciones recientes" />
      <List dense={true}>
        {entries.map((entry) => {
          return generateListItem(entry);
        })}
      </List>
    </Card>
  );
};

export default HistoryEntriesList;
