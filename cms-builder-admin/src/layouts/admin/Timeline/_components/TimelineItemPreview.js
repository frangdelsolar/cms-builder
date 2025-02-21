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
  Typography, // Import Typography
} from "@mui/material";
import Grid from "@mui/material/Grid2";
import { useAppSelector } from "../../../../store/Hooks";
import { selectSelectedEntity } from "../../../../store/EntitySlice";
import SearchIcon from "@mui/icons-material/Search";
import { useEffect, useState } from "react";
import { useContext } from "react";
import { ApiContext } from "../../../../context/ApiContext";
import { useNotifications } from "../../../../context/ToastContext";

import { useTheme } from "@mui/material/styles";
import MobileStepper from "@mui/material/MobileStepper";
import Button from "@mui/material/Button";
import KeyboardArrowLeft from "@mui/icons-material/KeyboardArrowLeft";
import KeyboardArrowRight from "@mui/icons-material/KeyboardArrowRight";

function TimelineItemPreview() {
  const entity = useAppSelector(selectSelectedEntity);
  const apiService = useContext(ApiContext);
  const toast = useNotifications();

  const [resourceId, setResourceId] = useState(null);
  const [timeline, setTimeline] = useState(null);
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 10,
    total: 0,
  });
  const [initialState, setInitialState] = useState(null);
  const [currentState, setCurrentState] = useState(null);
  const [currentEvent, setCurrentEvent] = useState(null);

  const handleResourceIdInputClick = async () => {
    try {
      const res = await apiService.getTimelineForResource(
        resourceId,
        entity.name,
        pagination.pageSize,
        pagination.page
      );

      setTimeline(res.data);
      setPagination(res.pagination);
      if (res.data.length > 0) {
        const state = JSON.parse(res.data[0].detail);
        setInitialState(state);
        setCurrentState(state);
        setCurrentEvent(res.data[0]); // Set the initial event
      } else {
        toast.show("ResourceId has no timeline", "info");
        setCurrentState(null);
        setCurrentEvent(null);
        setInitialState(null);
        setTimeline(null);
        setPagination({ page: 1, pageSize: 10, total: 0 });
      }
    } catch (error) {
      toast.show(`Error fetching timeline: ${error.message}`, "error");
    }
  };

  const [activeStep, setActiveStep] = useState(0);

  const lookAt = (stepNumber) => {
    const nextActiveStep = stepNumber;
    const step = timeline[stepNumber];
    const newState = applyChanges(initialState, stepNumber, timeline);
    setCurrentState(newState);
    setCurrentEvent(step);
    setActiveStep(nextActiveStep);
  };

  if (!entity) {
    return <></>;
  }

  return (
    <Card>
      <CardHeader title={`Timeline for ${entity.name}`}></CardHeader>
      <CardContent>
        <Grid container direction="column" spacing={2}>
          <ResourceIdInput
            resourceId={resourceId}
            setResourceId={setResourceId}
            onClick={handleResourceIdInputClick}
          />
          <ActionLabel event={currentEvent} />

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

const applyChanges = (initialState, stepNumber, timeline) => {
  let state = initialState;
  for (let i = 0; i < stepNumber + 1; i++) {
    const step = timeline[i];
    if (step.detail == "") {
      return null;
    }
    const changes = JSON.parse(step.detail);
    state = forwardChanges(state, changes);
  }
  return state;
};

function forwardChanges(originalObject, changes) {
  if (!originalObject) {
    return changes ? { ...changes } : {}; // Handle null originalObject
  }
  if (!changes) {
    return originalObject; // Handle null changes
  }

  const newObject = { ...originalObject };

  for (const key in changes) {
    if (changes.hasOwnProperty(key)) {
      const changeValue = changes[key];

      if (Array.isArray(changeValue) && changeValue.length === 2) {
        // [before, after] format
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
          }); // Recursive call for nested objects. Pass in the nested object and the change {key: afterValue}
        } else {
          newObject[key] = afterValue; // Directly replace with the after value
        }
      } else {
        newObject[key] = changeValue; // Handle other change formats if needed.
      }
    }
  }

  return newObject;
}

const EventPreview = ({ event }) => {
  if (!event || !event.detail) {
    return null;
  }

  const data = JSON.parse(event.detail);

  return <ObjectPreview jsonData={data} title="Introduced changes" />;
};

const ObjectPreview = ({ jsonData, title }) => {
  if (!jsonData) {
    return null;
  }

  return (
    <Paper
      elevation={0}
      sx={{
        padding: 2,
        backgroundColor: "#f5f5f5",
        overflow: "auto", // Enables scrolling if needed
        maxWidth: "100%", // Make sure it takes the full width
      }}
    >
      <Typography variant="h6">{title}</Typography>
      <pre
        style={{
          whiteSpace: "pre-wrap", // or 'pre-line' for wrapping
          wordBreak: "break-word", // For long words
          maxWidth: "100%",
          overflow: "auto", // Enables scrolling if needed
          maxHeight: "400px", // Optional: Set a max height for scrolling
        }}
      >
        {JSON.stringify(jsonData, null, 2)}
      </pre>
    </Paper>
  );
};

const ActionLabel = ({ event }) => {
  if (!event) {
    return null;
  }

  const formattedTime = new Date(event.UpdatedAt).toLocaleString(); // Format the time

  return (
    <Paper
      elevation={0}
      sx={{
        padding: 2,
        marginTop: 2,
        backgroundColor: "#f5f5f5",
        width: "100%",
        overflowWrap: "break-word",
        wordWrap: "break-word",
        hyphens: "auto",
      }}
    >
      <Typography
        sx={{
          display: "pre-wrap",
          wordBreak: "break-word",
        }}
      >
        <strong style={{ marginRight: "5px" }}>{event.username}</strong>{" "}
        {event.action} {event.resourceName} {event.resourceId} at{" "}
        {formattedTime}
      </Typography>
    </Paper>
  );
};

const Stepper = ({ pagination, timeline, lookAt, activeStep }) => {
  const theme = useTheme();
  if (!timeline) {
    return null;
  }

  return (
    <MobileStepper
      variant="dots"
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
          Next
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
          )}
          Back
        </Button>
      }
    />
  );
};

const ResourceIdInput = ({ resourceId, setResourceId, onClick }) => {
  return (
    <FormControl
      fullWidth
      variant="outlined"
      value={resourceId}
      onChange={(event) => {
        setResourceId(event.target.value);
      }}
    >
      <InputLabel htmlFor="outlined-input">Resource Id</InputLabel>
      <OutlinedInput
        id="outlined-input"
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
