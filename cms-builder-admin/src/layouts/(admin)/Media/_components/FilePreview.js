import { useEffect, useState, useContext } from "react";
import {
  Typography,
  Paper,
  Card,
  CardHeader,
  CardContent,
  CardActions,
  Button,
} from "@mui/material";
import Grid from "@mui/material/Grid2";

import { ApiContext } from "../../../../context/ApiContext";
import { useNotifications } from "../../../../context/ToastContext";
import { useDialogs } from "../../../../context/DialogContext";

const FilePreview = (props) => {
  const file = props.file;

  const apiService = useContext(ApiContext);
  const toast = useNotifications();
  const dialogs = useDialogs();

  const formatDate = (dateString) => {
    return dateString ? new Date(dateString).toLocaleString() : "N/A";
  };

  const handleDownload = async () => {
    try {
      const response = await apiService.downloadFile(file.ID);
      const blob = new Blob([response], { type: "application/octet-stream" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = file.name;
      link.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error("Error downloading file:", error);
      toast.show("Error downloading file", "error");
    }
  };

  const handleDelete = async () => {
    const confirmed = await dialogs.confirm({
      content: "¿Desea eliminar el archivo?",
    });
    if (!confirmed) return;

    try {
      await apiService.destroy("files", file.ID);
      toast.show("Item deleted successfully", "success");
      window.location.reload(); // TODO: Consider a more targeted update if possible
    } catch (error) {
      console.error("Error deleting item:", error);
      toast.show("Error deleting file", error.message);
    }
  };

  if (!file) return null; // Simplified condition

  return (
    <Card>
      <CardHeader title="File Info" />
      <CardContent>
        <Paper elevation={0} sx={{ padding: 2 }}>
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6}>
              <Typography variant="subtitle1">Name:</Typography>
              <Typography>{file.name || "N/A"}</Typography>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Typography variant="subtitle1">Size:</Typography>
              <Typography>
                {file.size ? `${file.size} bytes` : "N/A"}
              </Typography>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Typography variant="subtitle1">Last Modified:</Typography>
              <Typography>{formatDate(file.UpdatedAt)}</Typography>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Typography variant="subtitle1">Content Type:</Typography>
              <Typography>{file.mimeType || "N/A"}</Typography>
            </Grid>
          </Grid>
        </Paper>
      </CardContent>
      <CardActions sx={{ display: "flex", justifyContent: "flex-end" }}>
        <Button size="small" color="error" onClick={handleDelete}>
          Delete
        </Button>
        <Button size="small" color="primary" onClick={handleDownload}>
          Download
        </Button>
      </CardActions>
    </Card>
  );
};

export default FilePreview;
