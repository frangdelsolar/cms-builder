import { useEffect, useState, useContext } from "react";
import { useAppDispatch, useAppSelector } from "../../../../store/Hooks";

import { capitalize } from "@mui/material";
import Card from "@mui/material/Card";
import CardHeader from "@mui/material/CardHeader";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemText from "@mui/material/ListItemText";

import {
  selectEntities,
  setEntities,
  setSelectedEntity,
} from "../../../../store/EntitySlice";
import { ApiContext } from "../../../../context/ApiContext";

function ModelList() {
  const dispatch = useAppDispatch();
  const apiService = useContext(ApiContext);
  const entities = useAppSelector(selectEntities);
  const selectedEntity = useAppSelector((state) => state.entity.selectedEntity);

  const [listItems, setListItems] = useState([]);

  useEffect(() => {
    let getModels = async () => {
      try {
        let response = await apiService.getEntities();
        let data = response.data;
        data.sort((a, b) => a.pluralName.localeCompare(b.pluralName));
        dispatch(setEntities(data));
      } catch (error) {
        // TODO: have a toaster notification
        alert(error);
      }
    };

    getModels();
  }, []);

  const onEntityClick = (entity) => {
    dispatch(setSelectedEntity(entity));
    // add the selected entity to the query params
    window.location.href = `/models#${entity.kebabPluralName}`;
  };

  useEffect(() => {
    if (!entities) {
      return;
    }

    var items = [];
    let entitiesCopy = [...entities];
    entitiesCopy.forEach((entity) => {
      items.push(
        <ListItem disablePadding key={entity.pluralName}>
          <ListItemButton
            selected={selectedEntity === entity}
            onClick={() => onEntityClick(entity)}
          >
            <ListItemText primary={capitalize(entity.pluralName)} />
          </ListItemButton>
        </ListItem>
      );
    });

    setListItems(items);
  }, [entities, selectedEntity]);

  return (
    <Card>
      <CardHeader title="Recursos" />
      <List component="nav" aria-labelledby="nested-list-subheader">
        {listItems}
      </List>
    </Card>
  );
}

export default ModelList;
