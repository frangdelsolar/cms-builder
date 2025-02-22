import { useContext, useEffect, useState } from "react";
import {
  Typography,
  Button,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from "@mui/material";
import { ExpandMore } from "@mui/icons-material";
import SyncIcon from "@mui/icons-material/Sync";
import { ApiContext } from "../../../../context/ApiContext";
import { useNotifications } from "../../../../context/ToastContext";
import { formatChanges } from "./utils";

const RequestPreview = ({ requestId }) => {
  const apiService = useContext(ApiContext);
  const toast = useNotifications();
  const [isLoading, setIsLoading] = useState(false);
  const [requestDetails, setRequestDetails] = useState(null);
  const [expanded, setExpanded] = useState(false);

  useEffect(() => {
    if (requestId) {
      handleFetchClick();
    } else {
      resetState();
    }
  }, [requestId]);

  const resetState = () => {
    setRequestDetails(null);
    setExpanded(false);
  };

  const handleFetchClick = async () => {
    setIsLoading(true);
    try {
      const resp = await apiService.getRequestLogEntries(requestId);

      const actions = [];
      resp.data.history_entries.forEach((action) => {
        const label = `${action.username} ${action.action} ${action.resourceName} (${action.resourceId})`;
        actions.push(label);
      });

      let data = {
        ...resp.data.request_log,
        actions: actions,
      };

      const formattedData = formatChanges(data);
      setRequestDetails(formattedData);
      setExpanded(false);
    } catch (error) {
      toast.show("There was an error fetching the request log", "error");
    } finally {
      setIsLoading(false);
    }
  };

  const handleChange = (panel) => (requestId, isExpanded) => {
    setExpanded(isExpanded ? panel : false);
  };

  if (!requestId) return null;

  return (
    <Accordion
      expanded={expanded}
      onChange={handleChange("panel1")}
      elevation={0}
      sx={{ backgroundColor: "#f5f5f5", width: "100%" }}
    >
      <AccordionSummary expandIcon={<ExpandMore />}>
        <Typography sx={{ display: "pre-wrap", wordBreak: "break-word" }}>
          <strong style={{ marginRight: "5px" }}>Request Id:</strong>
          {requestId}
        </Typography>
      </AccordionSummary>
      <AccordionDetails>
        <pre
          style={{
            whiteSpace: "pre-wrap",
            wordBreak: "break-word",
          }}
        >
          {requestDetails ? (
            requestDetails
          ) : (
            <Button
              loading={isLoading}
              variant="outlined"
              onClick={handleFetchClick}
              loadingPosition="end"
              startIcon={<SyncIcon />}
            >
              Fetch
            </Button>
          )}
        </pre>
      </AccordionDetails>
    </Accordion>
  );
};

export default RequestPreview;
