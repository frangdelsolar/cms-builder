import React, { useContext, useEffect, useState, useMemo } from "react";
import {
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Typography,
  Paper,
  Stack,
  Chip,
  Button,
  CircularProgress,
  Card,
  CardContent,
  CardHeader,
} from "@mui/material";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import { CheckCircle, Error } from "@mui/icons-material";
import SyncIcon from "@mui/icons-material/Sync";
import { ApiContext } from "../../../../context/ApiContext";
import { useNotifications } from "../../../../context/ToastContext";

function SchedulerWidget() {
  const apiService = useContext(ApiContext);
  const toast = useNotifications();

  const [tasks, setTasks] = useState([]); // Initialize as an empty array
  const [isLoading, setIsLoading] = useState(false);

  // Fetch initial tasks
  useEffect(() => {
    const fetchTasks = async () => {
      setIsLoading(true);
      try {
        const res = await apiService.list("scheduler-tasks", 1, 10);
        setTasks(res.data); // Use an empty array as fallback
      } catch (error) {
        toast.show(`Error fetching tasks: ${error.message}`, "error");
      } finally {
        setIsLoading(false);
      }
    };

    fetchTasks();
  }, []);

  return (
    <Card>
      <CardHeader title="Scheduler" />
      <CardContent>
        {isLoading ? (
          <CircularProgress />
        ) : (
          <TaskAccordion tasks={tasks} setTasks={setTasks} />
        )}
      </CardContent>
    </Card>
  );
}

const TaskAccordion = ({ tasks, setTasks }) => {
  const apiService = useContext(ApiContext);
  const toast = useNotifications();

  const [isRunning, setIsRunning] = useState(false);

  // Group tasks by jobDefinitionName
  const groupedTasks = useMemo(() => {
    if (!Array.isArray(tasks)) {
      return {}; // Return an empty object if tasks is not an array
    }
    return tasks.reduce((acc, task) => {
      const key = task.jobDefinitionName;
      if (!acc[key]) {
        acc[key] = [];
      }
      acc[key].push(task);
      return acc;
    }, {});
  }, [tasks]);

  // Run a job and update the task list
  const runJob = async (jobName) => {
    setIsRunning(true);
    try {
      const res = await apiService.runJob(jobName);

      if (res.success) {
        toast.show(res.message, "success");
        // Add the new task to the tasks array
        setTasks((prevTasks) => [res.data, ...prevTasks]);
      }
    } catch (error) {
      toast.show(`Error running job: ${error.message}`, "error");
    } finally {
      setIsRunning(false);
    }
  };

  return (
    <Stack spacing={2}>
      {Object.entries(groupedTasks).map(([jobName, runs]) => {
        const successCount = runs.filter((run) => run.status === "done").length;
        const failedCount = runs.filter(
          (run) => run.status === "failed"
        ).length;

        return (
          <Accordion
            key={jobName}
            elevation={0}
            sx={{ backgroundColor: "#f5f5f5" }}
          >
            <AccordionSummary expandIcon={<ExpandMoreIcon />}>
              <Typography variant="h6" gutterBottom>
                {jobName}
              </Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Paper
                elevation={0}
                sx={{ padding: 2, width: "100%", backgroundColor: "#f5f5f5" }}
              >
                <Stack direction="row" spacing={2} alignItems="center">
                  <Typography variant="subtitle1" gutterBottom>
                    Total Runs: {runs.length} | Success: {successCount} |
                    Failed: {failedCount}
                  </Typography>
                  <Button
                    variant="outlined"
                    onClick={() => runJob(jobName)}
                    disabled={isRunning}
                    startIcon={<SyncIcon />}
                  >
                    {isRunning ? "Running..." : "Run"}
                  </Button>
                </Stack>

                {runs.map((run) => (
                  <Accordion key={run.ID} sx={{ marginTop: 2 }}>
                    <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                      <Stack direction="row" spacing={2} alignItems="center">
                        {run.status === "done" ? (
                          <CheckCircle color="success" />
                        ) : (
                          <Error color="error" />
                        )}
                        <Typography>Run {run.ID}</Typography>
                        <Chip
                          label={new Date(run.CreatedAt).toLocaleString()}
                          size="small"
                        />
                      </Stack>
                    </AccordionSummary>
                    <AccordionDetails>
                      <Stack spacing={1}>
                        <Typography>
                          <strong>Created At:</strong>{" "}
                          {new Date(run.CreatedAt).toLocaleString()}
                        </Typography>
                        <Typography>
                          <strong>Updated At:</strong>{" "}
                          {new Date(run.UpdatedAt).toLocaleString()}
                        </Typography>
                        {run.error && (
                          <Typography>
                            <strong>Error:</strong> {run.error}
                          </Typography>
                        )}
                        {run.results && (
                          <Typography component="div">
                            <strong>Results:</strong>
                            <pre
                              style={{
                                whiteSpace: "pre-wrap",
                                wordBreak: "break-word",
                              }}
                            >
                              {run.results}
                            </pre>
                          </Typography>
                        )}
                      </Stack>
                    </AccordionDetails>
                  </Accordion>
                ))}
              </Paper>
            </AccordionDetails>
          </Accordion>
        );
      })}
    </Stack>
  );
};

export default SchedulerWidget;
