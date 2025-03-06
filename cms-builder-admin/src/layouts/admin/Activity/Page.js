import Grid from "@mui/material/Grid2";
import HistoryEntriesList from "./_components/HistoryEntriesList";
import { useContext, useEffect, useState, useRef } from "react";
import { ApiContext } from "../../../context/ApiContext";
import { useNotifications } from "../../../context/ToastContext";
import ResearchRequests from "./_components/ResearchRequests";
import MostActiveUsers from "./_components/MostActiveUsers";
import {
  LocalLineChart,
  LocalPieChart,
  prepareLineChartData,
} from "./_components/Charts";
import ApiLatencyBarChart from "./_components/ApiLatencyBarChart";
import SchedulerWidget from "./_components/SchedulerWidget";

const AGGREGATION_INTERVAL = 20; // in minutes

export default function ActivityPage() {
  const apiService = useContext(ApiContext);
  const toast = useNotifications();
  const [stats, setStats] = useState([]);
  const isMounted = useRef(false);

  const [formattedData, setFormattedData] = useState({
    method_groups: [],
    status_groups: [],
  });

  useEffect(() => {
    if (isMounted.current) return;
    const getErrorRequests = async () => {
      try {
        const response = await apiService.getRequestStats();
        setStats(response.data);
      } catch (error) {
        console.log(error);
        toast.show("Error fetching error requests", "error");
      }
    };

    getErrorRequests();
    isMounted.current = true;
  }, []);

  useEffect(() => {
    if (!stats || stats.length === 0) {
      return;
    }

    const formattedData = {
      method_groups: stats.method_groups.map((stat) => ({
        label: stat.method,
        value: stat.count,
      })),
      status_groups: stats.status_groups.map((stat) => ({
        label: stat.status_code,
        value: stat.count,
      })),
      endpoints: stats.endpoints.map((stat) => ({
        label: stat.path,
        value: stat.count,
      })),
      requests: prepareLineChartData(stats.requests, AGGREGATION_INTERVAL),
    };

    setFormattedData(formattedData);
  }, [stats]);

  return (
    <Grid>
      <Grid container padding={1} spacing={2}>
        <Grid
          container
          size={{ xs: 12, sm: 3, md: 3 }}
          style={{ height: "100%" }}
          spacing={2}
        >
          <Grid size={{ xs: 12 }}>
            <ResearchRequests data={stats.requests} />
          </Grid>
          <Grid size={{ xs: 12 }}>
            <HistoryEntriesList />
          </Grid>
        </Grid>
        <Grid size="grow">
          <Grid container spacing={2}>
            <Grid xs={12} style={{ width: "100%" }}>
              <LocalLineChart
                title={`Requests through time every ${AGGREGATION_INTERVAL} mins`}
                data={formattedData.requests}
              />
            </Grid>
            <Grid size={{ xs: 12, md: 4 }}>
              <MostActiveUsers users={stats.users} />
            </Grid>
            <Grid size={{ xs: 12, md: 4 }}>
              <LocalPieChart
                title="Request Methods"
                data={formattedData.method_groups}
              />
            </Grid>
            <Grid size={{ xs: 12, md: 4 }}>
              <LocalPieChart
                title="Request Status"
                data={formattedData.status_groups}
              />
            </Grid>
            <Grid size={{ xs: 12, md: 6 }}>
              <ApiLatencyBarChart data={stats.requests} />
            </Grid>
            <Grid size={{ xs: 12, md: 6 }}>
              <SchedulerWidget />
            </Grid>
          </Grid>
        </Grid>
      </Grid>
    </Grid>
  );
}
