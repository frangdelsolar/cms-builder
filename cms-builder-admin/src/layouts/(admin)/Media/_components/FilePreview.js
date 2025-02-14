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
  const apiService = useContext(ApiContext);
  const [fileInfo, setFileInfo] = useState({});
  const toast = useNotifications();
  const dialogs = useDialogs();

  const fetchFileInfo = async (file) => {
    try {
      const response = await apiService.getFileInfo(file);
      setFileInfo(response.data);
    } catch (error) {
      console.error("Error fetching file info:", error);
      toast.show("Error fetching file info", "error");
    }
  };

  useEffect(() => {
    if (props.file && props.file.children.length <= 0) {
      fetchFileInfo(props.file.path);
    }
  }, [props.file]);

  const formatDate = (dateString) => {
    return dateString ? new Date(dateString).toLocaleString() : "N/A";
  };

  const handleDownload = async () => {
    try {
      const response = await apiService.downloadFile(props.file.path);
      const blob = new Blob([response], { type: "application/octet-stream" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = props.file.label;
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
      await apiService.deleteFile(props.file.path);
      toast.show("Item deleted successfully", "success");
      window.location.reload(); // TODO: Consider a more targeted update if possible
    } catch (error) {
      console.error("Error deleting item:", error);
      toast.show("Error deleting file", "error");
    }
  };

  if (!props.file) return null; // Simplified condition

  return (
    <Card>
      <CardHeader title="File Info" />
      <CardContent>
        {Object.keys(fileInfo).length > 0 ? (
          <Paper elevation={0} sx={{ padding: 2 }}>
            <Grid container spacing={2}>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle1">Name:</Typography>
                <Typography>{props.file.label || "N/A"}</Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle1">Size:</Typography>
                <Typography>
                  {fileInfo.size ? `${fileInfo.size} bytes` : "N/A"}
                </Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle1">Last Modified:</Typography>
                <Typography>{formatDate(fileInfo.last_modified)}</Typography>
              </Grid>
              <Grid item xs={12} sm={6}>
                <Typography variant="subtitle1">Content Type:</Typography>
                <Typography>{fileInfo.content_type || "N/A"}</Typography>
              </Grid>
            </Grid>
          </Paper>
        ) : (
          <Typography>No file information available.</Typography>
        )}
      </CardContent>
      <CardActions sx={{ display: "flex", justifyContent: "flex-end" }}>
        {Object.keys(fileInfo).length > 0 && (
          <>
            <Button size="small" color="error" onClick={handleDelete}>
              Delete
            </Button>
            <Button size="small" color="primary" onClick={handleDownload}>
              Download
            </Button>
          </>
        )}
      </CardActions>
    </Card>
  );
};

export default FilePreview;
