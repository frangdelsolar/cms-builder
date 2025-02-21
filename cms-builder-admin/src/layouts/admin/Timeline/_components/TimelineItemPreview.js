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

  const [currentState, setCurrentState] = useState(null);
  const [currentEvent, setCurrentEvent] = useState(null); // Store the current event details

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
        setCurrentState(JSON.parse(res.data[0].detail));
        setCurrentEvent(res.data[0]); // Set the initial event
      } else {
        toast.show("ResourceId has no timeline", "info");
        setCurrentState(null);
        setCurrentEvent(null);
      }
    } catch (error) {
      toast.show(`Error fetching timeline: ${error.message}`, "error");
    }
  };

  // Stepper

  const [activeStep, setActiveStep] = useState(0);

  useEffect(() => {
    if (!timeline) {
      return;
    }

    const step = timeline[activeStep];

    if (!step || !step.detail) {
      return;
    }
    try {
      const state = JSON.parse(step.detail);
      setCurrentState(state);
      setCurrentEvent(step); // Update current event details
    } catch (error) {
      console.error("Error parsing details:", error, step.detail);
      toast.show(`Error parsing details: ${error.message}`, "error");
    }
  }, [activeStep, timeline]); // Add timeline to the dependency array

  const ObjectPreview = ({ props }) => {
    return (
      <Paper
        elevation={0}
        sx={{
          padding: 2,
          marginTop: 2,
          backgroundColor: "#f5f5f5",
          overflow: "auto",
        }}
      >
        {currentState && <pre>{JSON.stringify(currentState, null, 2)}</pre>}
      </Paper>
    );
  };

  const EventDetails = ({ event }) => (
    <Paper
      elevation={0}
      sx={{
        padding: 2,
        marginTop: 2,
        backgroundColor: "#f5f5f5",
        overflow: "auto",
      }}
    ></Paper>
  );

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

          <Grid item container direction="row" spacing={2} sx={{ flexGrow: 1 }}>
            <Grid item xs={12} sm={6} sx={{ flexGrow: 1 }}>
              <EventDetails />
            </Grid>
            <Grid item xs={12} sm={6} sx={{ flexGrow: 1 }}>
              <ObjectPreview />
            </Grid>
          </Grid>

          <Stepper
            pagination={pagination}
            activeStep={activeStep}
            setActiveStep={setActiveStep}
          />
        </Grid>
      </CardContent>
    </Card>
  );
}

export default TimelineItemPreview;

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

const Stepper = ({ pagination, activeStep, setActiveStep }) => {
  const theme = useTheme();

  const handleNext = () => {
    setActiveStep((prevActiveStep) => prevActiveStep + 1);
  };

  const handleBack = () => {
    setActiveStep((prevActiveStep) => prevActiveStep - 1);
  };
  return (
    <MobileStepper
      variant="dots"
      steps={pagination.total}
      position="static"
      activeStep={activeStep}
      sx={{ flexGrow: 1 }}
      nextButton={
        <Button size="small" onClick={handleNext} disabled={activeStep === 5}>
          Next
          {theme.direction === "rtl" ? (
            <KeyboardArrowLeft />
          ) : (
            <KeyboardArrowRight />
          )}
        </Button>
      }
      backButton={
        <Button size="small" onClick={handleBack} disabled={activeStep === 0}>
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
