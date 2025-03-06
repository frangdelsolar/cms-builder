import { useEffect, useState, useContext, useRef } from "react";
import {
  Typography,
  Paper,
  Card,
  CardHeader,
  CardContent,
  CardActions,
  Button,
  IconButton,
  Tooltip,
  Stack,
  Box,
} from "@mui/material";
import Grid from "@mui/material/Grid2";
import FileCopyIcon from "@mui/icons-material/FileCopy";
import DownloadIcon from "@mui/icons-material/Download";
import { QRCodeCanvas } from "qrcode.react";

import { ApiContext } from "../../../../context/ApiContext";
import { useNotifications } from "../../../../context/ToastContext";
import { useDialogs } from "../../../../context/DialogContext";
import axios from "axios";

const FilePreview = (props) => {
  const file = props.file;

  const apiService = useContext(ApiContext);
  const toast = useNotifications();
  const dialogs = useDialogs();

  const formatDate = (dateString) => {
    return dateString ? new Date(dateString).toLocaleString() : "N/A";
  };

  const fileUrlRef = useRef(null);
  const privateFileUrlRef = useRef(null);

  const [privateDownloadUrl, setPrivateDownloadUrl] = useState(null);

  useEffect(() => {
    if (file) {
      setPrivateDownloadUrl(
        `${apiService.apiUrl()}/private/api/files/${file.ID}/download`
      );
    }
  }, [file]);

  const handleCopyUrl = (ref) => {
    if (ref.current) {
      navigator.clipboard
        .writeText(ref.current.innerText)
        .then(() => {
          toast.show("URL copied to clipboard", "success");
        })
        .catch((error) => {
          console.error("Failed to copy URL:", error);
          toast.show("Failed to copy URL", "error");
        });
    }
  };

  const handleDownload = async () => {
    try {
      const response = await apiService.downloadFile(file.ID);

      // Create a URL for the blob
      const url = URL.createObjectURL(response.data);

      // Create a temporary anchor element to trigger the download
      const link = document.createElement("a");
      link.href = url;
      link.download = file.name;
      document.body.appendChild(link);
      link.click();

      // Clean up
      document.body.removeChild(link);
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
      props.refresh();
    } catch (error) {
      console.error("Error deleting item:", error);
      toast.show("Error deleting file", error.message);
    }
  };

  if (!file) return null;

  return (
    <Card>
      <CardHeader
        title={
          <Typography variant="h5" style={{ wordWrap: "break-word" }}>
            {file.name || "N/A"}
          </Typography>
        }
        avatar={
          <Box>
            <QRCodeCanvas value={privateDownloadUrl} size={70} />
          </Box>
        }
      />
      <CardContent>
        <Paper elevation={0}>
          <Grid container spacing={2}>
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

        <Stack spacing={2} sx={{ marginTop: 2 }}>
          <Typography variant="h6">Private access</Typography>
          <Paper
            elevation={0}
            sx={{
              padding: 2,
              backgroundColor: "#f5f5f5",
              display: "flex",
              alignItems: "center",
            }}
          >
            <Typography
              variant="p"
              fontFamily={"monospace"}
              ref={privateFileUrlRef}
              sx={{
                flexGrow: 1,
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              }}
            >
              {privateDownloadUrl}
            </Typography>
            <Tooltip title="Copy URL">
              <IconButton
                onClick={() => handleCopyUrl(privateFileUrlRef)}
                aria-label="copy"
              >
                <FileCopyIcon />
              </IconButton>
            </Tooltip>
            <Tooltip title="Download">
              <IconButton onClick={() => handleDownload} aria-label="download">
                <DownloadIcon />
              </IconButton>
            </Tooltip>
          </Paper>
        </Stack>
      </CardContent>
      <CardActions sx={{ display: "flex", justifyContent: "flex-end" }}>
        <Button size="small" color="error" onClick={handleDelete}>
          Delete
        </Button>
      </CardActions>
    </Card>
  );
};

export default FilePreview;
