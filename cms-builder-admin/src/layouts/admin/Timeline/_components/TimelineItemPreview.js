import React, { useState, useContext, useEffect } from "react";
import {
  Card,
  CardHeader,
  CardContent,
  FormControl,
  InputLabel,
  OutlinedInput,
  InputAdornment,
  IconButton,
  Paper,
  Typography,
  MobileStepper,
  Button,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from "@mui/material";
import Grid from "@mui/material/Grid2";
import {
  KeyboardArrowLeft,
  KeyboardArrowRight,
  ExpandMore,
} from "@mui/icons-material";
import SearchIcon from "@mui/icons-material/Search";
import SyncIcon from "@mui/icons-material/Sync";
import { useAppSelector } from "../../../../store/Hooks";
import { selectSelectedEntity } from "../../../../store/EntitySlice";
import { ApiContext } from "../../../../context/ApiContext";
import { useNotifications } from "../../../../context/ToastContext";
import { useTheme } from "@mui/material/styles";

// Main Component
function TimelineItemPreview() {
  const entity = useAppSelector(selectSelectedEntity);
  const apiService = useContext(ApiContext);
  const toast = useNotifications();

  const [resourceId, setResourceId] = useState(0);
  const [timeline, setTimeline] = useState(null);
  const [pagination, setPagination] = useState({
    page: 1,
    limit: 10,
    total: 0,
  });
  const [initialState, setInitialState] = useState(null);
  const [currentState, setCurrentState] = useState(null);
  const [currentEvent, setCurrentEvent] = useState(null);
  const [activeStep, setActiveStep] = useState(0);
  const [lastVisitedPage, setLastVisitedPage] = useState(0);

  // Reset state when entity changes
  useEffect(() => {
    resetState();
  }, [entity?.name]);

  const handleResourceIdInputClick = async () => {
    const res = await getItems(1);
    setTimeline(res.data);

    if (res.data.length > 0) {
      const state = JSON.parse(res.data[0].detail);
      setInitialState(state);
      setCurrentState(state);
      setCurrentEvent(res.data[0]);
    } else {
      toast.show("ResourceId has no timeline", "info");
      resetState();
    }
  };

  const getItems = async (page) => {
    if (!resourceId > 0) return null;

    try {
      const res = await apiService.getTimelineForResource(
        resourceId,
        entity.name,
        pagination.limit,
        page
      );
      setPagination(res.pagination);
      setLastVisitedPage(res.pagination.page);
      return res;
    } catch (error) {
      toast.show(`Error fetching timeline: ${error.message}`, "error");
    }
  };

  const resetState = () => {
    setResourceId(0);
    setCurrentState(null);
    setCurrentEvent(null);
    setInitialState(null);
    setTimeline(null);
    setPagination({ page: 1, limit: 10, total: 0 });
  };

  const lookAt = async (stepNumber) => {
    const stepPage = Math.floor(stepNumber / pagination.limit) + 1;
    let newTimeline = [...timeline];
    if (stepPage > lastVisitedPage) {
      const res = await getItems(stepPage);
      newTimeline = [...newTimeline, ...res.data];
      setTimeline(newTimeline);
    }

    const newState = applyChanges(initialState, stepNumber, newTimeline);
    setCurrentState(newState);
    setCurrentEvent(newTimeline[stepNumber]);
    setActiveStep(stepNumber);
  };

  if (!entity) return null;

  return (
    <Card>
      <CardHeader title={`Timeline for ${entity.name}`} />
      <CardContent>
        <Grid container direction="column" spacing={2}>
          <ResourceIdInput
            key={entity?.name}
            resourceId={resourceId}
            setResourceId={setResourceId}
            onClick={handleResourceIdInputClick}
          />
          <ActionLabel event={currentEvent} />
          <RequestPreview event={currentEvent} />
          <Grid item container direction="row" spacing={2}>
            <Grid item xs={12} sm={6} sx={{ flexGrow: 1 }}>
              <EventPreview event={currentEvent} />
            </Grid>
            <Grid item xs={12} sm={6} sx={{ flexGrow: 1 }}>
              <ObjectPreview
                jsonData={currentState}
                title="State of the resource after implementing the changes"
              />
            </Grid>
          </Grid>
          <Stepper
            pagination={pagination}
            lookAt={lookAt}
            timeline={timeline}
            activeStep={activeStep}
          />
        </Grid>
      </CardContent>
    </Card>
  );
}

export default TimelineItemPreview;

// Utility Functions
const applyChanges = (initialState, stepNumber, timeline) => {
  let state = initialState;
  for (let i = 0; i < stepNumber + 1; i++) {
    const step = timeline[i];
    if (step.detail === "") return null;
    const changes = JSON.parse(step.detail);
    state = forwardChanges(state, changes);
  }
  return state;
};

const forwardChanges = (originalObject, changes) => {
  if (!originalObject) return changes ? { ...changes } : {};
  if (!changes) return originalObject;

  const newObject = { ...originalObject };

  for (const key in changes) {
    if (changes.hasOwnProperty(key)) {
      const changeValue = changes[key];
      if (Array.isArray(changeValue) && changeValue.length === 2) {
        const afterValue = changeValue[1];
        if (
          typeof afterValue === "object" &&
          afterValue !== null &&
          typeof newObject[key] === "object" &&
          newObject[key] !== null &&
          !Array.isArray(afterValue) &&
          !Array.isArray(newObject[key])
        ) {
          newObject[key] = forwardChanges(newObject[key], {
            [key]: afterValue,
          });
        } else {
          newObject[key] = afterValue;
        }
      } else {
        newObject[key] = changeValue;
      }
    }
  }
  return newObject;
};

const formatChanges = (changes) => {
  if (!changes || typeof changes !== "object") return "No changes";

  let formattedChanges = [];
  for (const key in changes) {
    if (changes.hasOwnProperty(key)) {
      const changeValue = changes[key];
      if (Array.isArray(changeValue) && changeValue.length === 2) {
        formattedChanges.push(
          `${key}: ${JSON.stringify(changeValue[0])} -> ${JSON.stringify(
            changeValue[1]
          )}`
        );
      } else if (typeof changeValue === "object" && changeValue !== null) {
        const nestedFormattedChanges = formatChanges(changeValue);
        if (nestedFormattedChanges) {
          formattedChanges.push(`${key}: { ${nestedFormattedChanges} }`);
        }
      } else {
        formattedChanges.push(`${key}: ${JSON.stringify(changeValue)}`);
      }
    }
  }
  return formattedChanges.join("\n");
};

// Sub-Components
const EventPreview = ({ event }) => {
  if (!event || !event.detail) return null;
  const data = formatChanges(JSON.parse(event.detail));
  return <ObjectPreview data={data} title="Introduced changes" />;
};

const ObjectPreview = ({ jsonData, title, data }) => {
  if (!jsonData && !data) return null;

  return (
    <Paper
      elevation={0}
      sx={{
        padding: 2,
        backgroundColor: "#f5f5f5",
        maxWidth: "100%",
        overflowWrap: "break-word",
      }}
    >
      <Typography variant="h6">{title}</Typography>
      <pre
        style={{
          whiteSpace: "pre-wrap",
          wordBreak: "break-word",
          overflow: "auto",
          maxHeight: "350px",
        }}
      >
        {jsonData && JSON.stringify(jsonData, null, 2)}
        {data && data}
      </pre>
    </Paper>
  );
};

const ActionLabel = ({ event }) => {
  if (!event) return null;
  const formattedTime = new Date(event.UpdatedAt).toLocaleString();

  return (
    <Paper
      elevation={0}
      sx={{
        padding: 2,
        backgroundColor: "#f5f5f5",
        width: "100%",
        overflowWrap: "break-word",
      }}
    >
      <Typography sx={{ display: "pre-wrap", wordBreak: "break-word" }}>
        <strong style={{ marginRight: "5px" }}>{event.username}</strong>{" "}
        {event.action.toUpperCase()} {event.resourceName} ({event.resourceId})
        at {formattedTime}
      </Typography>
    </Paper>
  );
};

const Stepper = ({ pagination, timeline, lookAt, activeStep }) => {
  const theme = useTheme();
  if (!timeline) return null;

  return (
    <MobileStepper
      variant="progress"
      steps={pagination.total}
      position="static"
      activeStep={activeStep}
      sx={{ flexGrow: 1 }}
      nextButton={
        <Button
          size="small"
          onClick={() => lookAt(activeStep + 1)}
          disabled={activeStep === pagination.total - 1}
        >
          Next{" "}
          {theme.direction === "rtl" ? (
            <KeyboardArrowLeft />
          ) : (
            <KeyboardArrowRight />
          )}
        </Button>
      }
      backButton={
        <Button
          size="small"
          onClick={() => lookAt(activeStep - 1)}
          disabled={activeStep === 0}
        >
          {theme.direction === "rtl" ? (
            <KeyboardArrowRight />
          ) : (
            <KeyboardArrowLeft />
          )}{" "}
          Back
        </Button>
      }
    />
  );
};

const ResourceIdInput = ({ resourceId, setResourceId, onClick }) => {
  return (
    <FormControl fullWidth variant="outlined">
      <InputLabel htmlFor="outlined-input">Resource Id</InputLabel>
      <OutlinedInput
        id="outlined-input"
        value={resourceId}
        onChange={(event) => setResourceId(event.target.value)}
        endAdornment={
          <InputAdornment position="end">
            <IconButton onClick={onClick} edge="end">
              <SearchIcon />
            </IconButton>
          </InputAdornment>
        }
        label="Resource Id"
      />
    </FormControl>
  );
};

const RequestPreview = ({ event }) => {
  const apiService = useContext(ApiContext);
  const toast = useNotifications();
  const [isLoading, setIsLoading] = useState(false);
  const [requestDetails, setRequestDetails] = useState(null);
  const [expanded, setExpanded] = useState(false);

  useEffect(() => {
    if (event?.requestId) {
      handleFetchClick();
    } else {
      resetState();
    }
  }, [event?.requestId]);

  const resetState = () => {
    setRequestDetails(null);
    setExpanded(false);
  };

  const handleFetchClick = async () => {
    setIsLoading(true);
    try {
      const resp = await apiService.getRequestLogEntries(event.requestId);
      const formattedData = formatChanges(resp.data);
      setRequestDetails(formattedData);
      setExpanded(false);
    } catch (error) {
      toast.show("There was an error fetching the request log", "error");
    } finally {
      setIsLoading(false);
    }
  };

  const handleChange = (panel) => (event, isExpanded) => {
    setExpanded(isExpanded ? panel : false);
  };

  if (!event || !event.requestId) return null;

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
          {event.requestId}
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
