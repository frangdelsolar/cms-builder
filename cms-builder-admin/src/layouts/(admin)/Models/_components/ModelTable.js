import { useEffect, useState, useContext, useCallback } from "react";
import { DataGrid, GridToolbar } from "@mui/x-data-grid";

import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import { capitalize, Typography } from "@mui/material";
import Button from "@mui/material/Button";

import Stack from "@mui/material/Stack";
import Skeleton from "@mui/material/Skeleton";

import ModelForm from "./ModelForm";

import { useDialogs } from "../../../../context/DialogContext";
import { useNotifications } from "../../../../context/ToastContext";
import { useAppDispatch, useAppSelector } from "../../../../store/Hooks";
import { ApiContext } from "../../../../context/ApiContext";

import {
  selectSelectedEntity,
  selectSchemas,
  setSchema,
} from "../../../../store/EntitySlice";
import { setFormSaving, setFormErrors } from "../../../../store/FormSlice";

import { getColumnsFromData, getRowsFromData } from "./utils";

const ModelTable = () => {
  const toast = useNotifications();
  const apiService = useContext(ApiContext);
  const dispatch = useAppDispatch();
  const entity = useAppSelector(selectSelectedEntity);
  const schemas = useAppSelector(selectSchemas);
  const dialogs = useDialogs();

  // Local state for the table
  const [isLoading, setIsLoading] = useState(true);
  const [paginationModel, setPaginationModel] = useState({
    page: 0,
    pageSize: 10,
  });
  const [total, setTotal] = useState(0);
  const [columns, setColumns] = useState(null);
  const [rows, setRows] = useState(null);
  const [sortModel, setSortModel] = useState([]);
  const [density, setDensity] = useState("compact");

  // Fetch schema for the selected entity
  useEffect(() => {
    if (!entity) return;

    setSortModel([]);

    const getSchemaForEntity = async () => {
      try {
        const response = await apiService.schema(entity.kebabPluralName);
        dispatch(setSchema({ key: entity.name, schema: response.data }));
        resetTable();
      } catch (error) {
        let errorMessage = "Error fetching schema: " + error.message;
        toast.show(errorMessage, {
          severity: "error",
        });
      }
    };

    getSchemaForEntity();
  }, [entity]);

  // Resets the table's state (pagination, columns, and rows)
  const resetTable = () => {
    setPaginationModel({ page: 0, pageSize: 10 });
    setTotal(0);
    setColumns(null);
    setRows(null);
  };

  // Handles row deletion with confirmation dialog
  const onClickDelete = (params) => async () => {
    const confirmed = await dialogs.confirm(
      "¿Desea eliminar el registro de " +
        capitalize(entity.name) +
        " " +
        params.row.ID +
        "?",
      {
        okText: "Sí",
        cancelText: "No",
      }
    );

    if (!confirmed) return;

    handleDelete(params);
  };

  const openModelFormDialog = (title, schema, data = null, submitHandler) => {
    const content = (
      <ModelForm schema={schema} data={data} submitHandler={submitHandler} />
    ); // Create content here
    const actions = [
      {
        label: "Guardar",
        onClick: () => {
          dispatch(setFormSaving(true));
        },
      },
    ];

    dialogs.show({
      title,
      content,
      actions,
    });
  };

  const onClickAdd = () => {
    const title = "Añadir Registro de " + capitalize(entity.name);
    openModelFormDialog(title, schemas[entity.name], null, handleSave); // Call common function
  };

  const onClickEdit = (params) => () => {
    const title = "Actualizar registro de " + capitalize(entity.name);
    openModelFormDialog(title, schemas[entity.name], params.row, handleEdit); // Call common function with data
  };

  // Generate columns for DataGrid based on the schema properties
  useEffect(() => {
    if (!schemas || !entity || !schemas[entity.name]) return;

    const entityRef = schemas[entity.name]["$ref"].split("/")[2];
    const properties = schemas[entity.name]["$defs"][entityRef]["properties"];
    const columns = getColumnsFromData(properties, onClickEdit, onClickDelete);
    setColumns(columns);
  }, [schemas, entity]);

  // Fetch rows for the selected entity based on pagination
  useEffect(() => {
    if (!entity || !schemas[entity.name]) return;

    setIsLoading(true);

    const getData = async () => {
      let sortBy = "";

      sortModel.forEach((option) => {
        if (option.sort == "desc") {
          sortBy = "-" + option.field + ",";
        } else {
          sortBy = option.field + ",";
        }
      });
      if (sortBy.length > 0) {
        sortBy = sortBy.slice(0, -1);
      }

      try {
        const response = await apiService.list(
          entity.kebabPluralName,
          paginationModel.page + 1,
          paginationModel.pageSize,
          sortBy
        );
        const rows = getRowsFromData(response.data, schemas[entity.name]);
        setRows(rows);
        setTotal(response.pagination.total);
        setIsLoading(false);
      } catch (error) {
        let errorMessage = "Error fetching data: " + error.message;
        toast.show(errorMessage, { severity: "error" });
      }
    };
    getData();
  }, [schemas, entity, paginationModel, sortModel]);

  // Defines actions (edit/delete) for each row

  // Handles the save action for editing
  const handleEdit = async (data) => {
    try {
      await apiService.put(entity.kebabPluralName, data, data);
      // TODO: Maybe I can update state directly
      toast.show("Item updated successfully", "success");
      window.location.reload();
    } catch (error) {
      // TODO: Show error message
      console.log("Error updating item: ", error);
      let errorMessage = "Error updating item: " + error.message;
      toast.show(errorMessage, "error");
      dispatch(setFormErrors(error.data.Errors));
    }
  };

  // Handles the save action for adding a new record
  const handleSave = async (data) => {
    try {
      await apiService.post(entity.kebabPluralName, data);
      dialogs.close();
      // TODO: Maybe I can update state directly
      toast.show("Item saved successfully", "success");
      window.location.reload();
    } catch (error) {
      let errorMessage = "Error saving item: " + error.message;
      toast.show(errorMessage, "error");
      dispatch(setFormErrors(error.data.Errors));
    }
  };

  // Handles the delete action
  const handleDelete = async (params) => {
    try {
      await apiService.destroy(entity.kebabPluralName, params.id);
      // TODO: Maybe I can update state directly
      toast.show("Item deleted successfully", "success");
      window.location.reload();
    } catch (error) {
      // TODO: Show error message
      let errorMessage = "Error deleting item: " + error.message;
      toast.show(errorMessage, "error");
    }
  };

  const handleSortModelChange = useCallback((sortModel) => {
    setSortModel(sortModel);
  }, []);

  // if (isLoading) {
  if (!schemas || !entity || !schemas[entity.name] || !columns || !rows) {
    return (
      <Card>
        <CardContent>
          <Stack direction="column" spacing={2} sx={{ mb: 1 }}>
            <Skeleton variant="rectangular" height={30} />
            <Skeleton variant="rectangular" height={300} />
          </Stack>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent>
        <Typography variant="h4" gutterBottom>
          {capitalize(entity.pluralName)}
        </Typography>
        <Stack direction="row" spacing={1} sx={{ mb: 1 }}>
          <Button size="small" onClick={onClickAdd}>
            Añadir registro
          </Button>
        </Stack>
        <DataGrid
          slots={{ toolbar: GridToolbar }}
          rows={rows}
          columns={columns}
          paginationModel={paginationModel}
          onPaginationModelChange={setPaginationModel}
          rowCount={total}
          pageSizeOptions={[5, 10, 25, 50, 100]}
          rowSelection={false}
          sx={{ border: 0 }}
          loading={isLoading}
          paginationMode="server"
          sortingMode="server"
          onSortModelChange={handleSortModelChange}
          density={density}
          onDensityChange={(newDensity) => setDensity(newDensity)}
        />
      </CardContent>
    </Card>
  );
};

export default ModelTable;
