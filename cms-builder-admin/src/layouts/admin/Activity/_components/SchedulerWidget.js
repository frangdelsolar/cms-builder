import React from "react";
import {
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Typography,
  Paper,
  Stack,
  Chip,
} from "@mui/material";
import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import { CheckCircle, Error } from "@mui/icons-material";
import { Card, CardContent, CardHeader } from "@mui/material";
import { useContext, useEffect, useState } from "react";
import { ApiContext } from "../../../../context/ApiContext";
import { useNotifications } from "../../../../context/ToastContext";

function SchedulerWidget() {
  const apiService = useContext(ApiContext);
  const toast = useNotifications();

  const [data, setData] = useState([]);

  useEffect(() => {
    const getData = async () => {
      try {
        const res = await apiService.list("scheduler-tasks", 1, 10);
        setData(res.data);
      } catch (error) {
        toast.show(`Error fetching scheduler tasks: ${error.message}`, "error");
      }
    };

    getData();
  }, []);

  return (
    <Card>
      <CardHeader title="Scheduler" />
      <CardContent>
        <TaskAccordion tasks={data} />
      </CardContent>
    </Card>
  );
}

export default SchedulerWidget;

const TaskAccordion = ({ tasks }) => {
  // Group tasks by jobDefinitionName
  const groupedTasks = tasks.reduce((acc, task) => {
    const key = task.jobDefinitionName;
    if (!acc[key]) {
      acc[key] = [];
    }
    acc[key].push(task);
    return acc;
  }, {});

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
                <Typography variant="subtitle1" gutterBottom>
                  Total Runs: {runs.length} | Success: {successCount} | Failed:{" "}
                  {failedCount}
                </Typography>
                {runs.map((run, index) => (
                  <Accordion key={run.ID} sx={{ marginTop: 2 }}>
                    <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                      <Stack direction="row" spacing={2} alignItems="center">
                        {run.status === "done" ? (
                          <CheckCircle color="success" />
                        ) : (
                          <Error color="error" />
                        )}
                        <Typography>
                          Run {index + 1} - {run.status}
                        </Typography>
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
