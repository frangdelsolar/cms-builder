import { GridActionsCellItem } from "@mui/x-data-grid";
import EditIcon from "@mui/icons-material/Edit";
import DeleteIcon from "@mui/icons-material/DeleteOutlined";
import { capitalize } from "@mui/material";

/**
 * Transforms raw data from the API into rows suitable for the DataGrid.
 * Handles type conversions based on the schema.
 *
 * @param {Array} data - The raw data from the API.
 * @param {Object} schema - The schema for the entity.
 * @returns {Array} - The transformed rows for the DataGrid.
 */
export function getRowsFromData(data, schema) {
  if (!schema || !schema["$ref"] || !schema["$defs"]) {
    console.error("Schema is missing $ref or $defs properties:", schema);
    return []; // Or throw an error, depending on your needs
  }

  const entityRef = schema["$ref"].split("/")[2];
  const properties = schema["$defs"][entityRef]["properties"];

  return data.map((item) => {
    let field = {};
    for (const [key, value] of Object.entries(item)) {
      if (key === "ID") {
        field["id"] = value;
        field["ID"] = value;
        continue;
      }

      const fieldSchema = properties[key];

      if (!fieldSchema) {
        console.warn(`No schema found for field: ${key}. Skipping.`);
        continue;
      }

      // Handle different data types
      switch (fieldSchema.type) {
        case "string":
          if (
            fieldSchema.format === "date-time" ||
            fieldSchema.format === "date"
          ) {
            field[key] = new Date(value);
          } else {
            field[key] = value;
          }
          break;
        case "number":
          field[key] = value;
          break;
        case "boolean": // Handle boolean type
          field[key] = value;
          break;
        default:
          field[key] = value;
      }
    }
    return field;
  });
}

/**
 * Generates DataGrid columns based on the schema properties.
 *
 * @param {Object} properties - The properties object from the schema.
 * @returns {Array} - The DataGrid column definitions.
 */
export function getColumnsFromData(properties, editHandler, deleteHandler) {
  const getColumnActions = (params) => [
    <GridActionsCellItem
      icon={<EditIcon />}
      label="Edit"
      className="textPrimary"
      onClick={editHandler(params)}
      color="inherit"
    />,
    <GridActionsCellItem
      icon={<DeleteIcon />}
      label="Delete"
      onClick={deleteHandler(params)}
      color="inherit"
    />,
  ];

  const columns = [];

  columns.push({
    field: "actions",
    type: "actions",
    headerName: "Actions",
    width: 100,
    cellClassName: "actions",
    getActions: getColumnActions, // This will be defined in your component
  });

  for (const key in properties) {
    let width = 150;
    if (key === "ID") {
      width = 30;
    } else if (key.toLowerCase().includes("id")) {
      // Improved check for "id"
      width = 100;
    }

    if (properties[key].hasOwnProperty("$ref")) {
      continue;
    }

    let fieldType = properties[key].type;
    if (properties[key].format === "date-time") {
      fieldType = "dateTime";
    } else if (properties[key].format === "date") {
      fieldType = "date";
    }

    columns.push({
      field: key,
      headerName: capitalize(key),
      width,
      align: "center",
      headerAlign: "center",
      type: fieldType,
    });
  }

  return columns;
}
