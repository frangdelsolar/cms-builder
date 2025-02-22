import {
  Card,
  CardContent,
  CardHeader,
  MenuItem,
  Select,
  Button,
  FormControl,
  InputLabel,
} from "@mui/material";
import Grid from "@mui/material/Grid2";
import { useContext, useEffect, useState, useRef } from "react";
import { ApiContext } from "../../../../context/ApiContext";
import { useNotifications } from "../../../../context/ToastContext";
import { PieChart } from "@mui/x-charts/PieChart";
import { LineChart } from "@mui/x-charts/LineChart";
import RequestPreview from "../../Timeline/_components/RequestPreview";

const AGGREGATION_INTERVAL = 50; // in minutes

function ErrorMonitor() {
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
      requests: prepareLineChartData(stats.requests),
    };

    setFormattedData(formattedData);
  }, [stats]);

  return (
    <Grid container spacing={2}>
      <Grid xs={12} style={{ width: "100%" }}>
        <LocalLineChart
          title={`Requests through time every ${AGGREGATION_INTERVAL} mins`}
          data={formattedData.requests}
        />
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
      <Grid size={{ xs: 12, md: 4 }}>
        <ResearchRequests data={stats.requests} />
      </Grid>
    </Grid>
  );
}

export default ErrorMonitor;

const ResearchRequests = ({ data }) => {
  const [statusCodes, setStatusCodes] = useState([]); // List of unique status codes
  const [selectedStatusCode, setSelectedStatusCode] = useState(""); // Selected status code
  const [requestIdentifiers, setRequestIdentifiers] = useState([]); // List of request identifiers for the selected status code
  const [selectedRequestIdentifier, setSelectedRequestIdentifier] =
    useState(""); // Selected request identifier
  const [selectedRequestDetail, setSelectedRequestDetail] = useState(null); // Details of the selected request

  // Extract unique status codes from the data
  useEffect(() => {
    if (data && Array.isArray(data)) {
      const uniqueStatusCodes = [
        ...new Set(data.map((request) => request.status_code)),
      ];
      setStatusCodes(uniqueStatusCodes);
    }
  }, [data]);

  // Update request identifiers when the selected status code changes
  useEffect(() => {
    if (selectedStatusCode && data && Array.isArray(data)) {
      const filteredRequests = data.filter(
        (request) => request.status_code === selectedStatusCode
      );
      const identifiers = filteredRequests.map(
        (request) => request.request_identifier
      );
      setRequestIdentifiers(identifiers);
      setSelectedRequestIdentifier(""); // Reset selected request identifier
    }
  }, [selectedStatusCode, data]);

  // Fetch request details when the button is clicked
  const fetchRequestDetail = () => {
    if (selectedRequestIdentifier && data && Array.isArray(data)) {
      const requestDetail = data.find(
        (request) => request.request_identifier === selectedRequestIdentifier
      );
      setSelectedRequestDetail(requestDetail);
    }
  };

  if (!data) {
    return null;
  }

  return (
    <Card>
      <CardHeader title="Researcher" />
      <CardContent>
        {/* Status Code Dropdown */}
        <FormControl
          fullWidth
          variant="outlined"
          style={{ marginBottom: "16px" }}
        >
          <InputLabel>Status Code</InputLabel>
          <Select
            value={selectedStatusCode}
            onChange={(e) => {
              setSelectedStatusCode(e.target.value);
              setSelectedRequestDetail("");
            }}
            label="Status Code"
          >
            {statusCodes.map((code) => (
              <MenuItem key={code} value={code}>
                {code}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        {/* Request Identifier Dropdown */}
        <FormControl
          fullWidth
          variant="outlined"
          style={{ marginBottom: "16px" }}
        >
          <InputLabel>Request Identifier</InputLabel>
          <Select
            value={selectedRequestIdentifier}
            onChange={(e) => {
              setSelectedRequestDetail("");
              setSelectedRequestIdentifier(e.target.value);
            }}
            label="Request Identifier"
            disabled={!selectedStatusCode} // Disable if no status code is selected
          >
            {requestIdentifiers.map((identifier) => (
              <MenuItem key={identifier} value={identifier}>
                {identifier}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        {/* Fetch Details Button */}
        <Button
          variant="contained"
          color="primary"
          onClick={fetchRequestDetail}
          disabled={!selectedRequestIdentifier} // Disable if no request identifier is selected
          style={{ marginBottom: "16px" }}
        >
          Fetch Request Detail
        </Button>

        {/* Display Request Details */}
        {selectedRequestDetail && (
          <RequestPreview
            requestId={selectedRequestDetail.request_identifier}
          />
        )}
      </CardContent>
    </Card>
  );
};

const LocalLineChart = ({ title, data }) => {
  if (!data) {
    return null;
  }

  return (
    <Card>
      <CardHeader title={title} />
      <CardContent>
        <LineChart
          series={data.series}
          xAxis={[{ scaleType: "point", data: data.xLabels }]}
          height={400}
        />
      </CardContent>
    </Card>
  );
};

function prepareLineChartData(logs) {
  if (!logs || !Array.isArray(logs)) {
    return { chartData: [], xLabels: [] };
  }

  const statusCodeData = {}; // Store data for each status code
  const timeIntervalData = {}; // Aggregate data every 60 minutes
  const blockSize = AGGREGATION_INTERVAL; // in minutes

  // Find the earliest and latest timestamps in the logs
  const timestamps = logs.map((log) => new Date(log.timestamp).getTime());
  const minTimestamp = Math.min(...timestamps);
  const maxTimestamp = Math.max(...timestamps);

  // Round the min and max timestamps to the nearest hour
  const startTime = new Date(
    Math.floor(minTimestamp / (blockSize * 60 * 1000)) * (blockSize * 60 * 1000)
  );
  const endTime = new Date(
    Math.ceil(maxTimestamp / (blockSize * 60 * 1000)) * (blockSize * 60 * 1000)
  );

  // Generate all hourly timestamps between startTime and endTime
  const allTimeKeys = [];
  for (
    let time = startTime;
    time <= endTime;
    time.setMinutes(time.getMinutes() + blockSize)
  ) {
    allTimeKeys.push(time.toISOString());
  }

  // Initialize timeIntervalData with all time keys
  allTimeKeys.forEach((timeKey) => {
    timeIntervalData[timeKey] = {};
  });

  // Process logs and populate statusCodeData and timeIntervalData
  logs.forEach((log) => {
    const statusCode = log.status_code;
    const timestamp = new Date(log.timestamp);

    // Round timestamp to the nearest blockSize minutes
    const roundedTimestamp = new Date(
      Math.round(timestamp.getTime() / (blockSize * 60 * 1000)) *
        (blockSize * 60 * 1000)
    );

    const timeKey = roundedTimestamp.toISOString();

    // Initialize counts for status codes
    if (!statusCodeData[statusCode]) {
      statusCodeData[statusCode] = {};
    }

    // Increment count for status code at this time interval
    statusCodeData[statusCode][timeKey] =
      (statusCodeData[statusCode][timeKey] || 0) + 1;
    timeIntervalData[timeKey][statusCode] =
      (timeIntervalData[timeKey][statusCode] || 0) + 1;
  });

  // Prepare data for LineChart
  const xLabels = allTimeKeys; // Use all generated time keys as labels

  const series = Object.keys(statusCodeData).map((statusCode) => {
    const data = xLabels.map((timeKey) => {
      return statusCodeData[statusCode][timeKey] || 0; // Fill in 0 for missing data
    });
    return { data, label: statusCode };
  });

  return { series, xLabels };
}

const LocalPieChart = ({ title, data }) => {
  if (!data) {
    return null;
  }

  return (
    <Card>
      <CardHeader title={title} />
      <CardContent style={{ height: "300px" }}>
        <PieChart
          series={[
            {
              data: data,
              innerRadius: 50,
              outerRadius: 100,
              highlightScope: { fade: "global", highlight: "item" },
              faded: { innerRadius: 30, additionalRadius: -30, color: "gray" },
            },
          ]}
        />
      </CardContent>
    </Card>
  );
};
