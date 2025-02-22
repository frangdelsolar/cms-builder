import { Card, CardContent, CardHeader } from "@mui/material";
import { PieChart } from "@mui/x-charts/PieChart";
import { LineChart } from "@mui/x-charts/LineChart";

function prepareLineChartData(logs, aggregationInterval) {
  if (!aggregationInterval) {
    console.error("Missing aggregation interval");
    return { chartData: [], xLabels: [] };
  }

  if (!logs || !Array.isArray(logs)) {
    return { chartData: [], xLabels: [] };
  }

  const statusCodeData = {}; // Store data for each status code
  const timeIntervalData = {}; // Aggregate data every 60 minutes
  const blockSize = aggregationInterval; // in minutes

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

    // Check if the timestamp is valid
    if (isNaN(timestamp.getTime())) {
      console.error("Invalid timestamp:", log.timestamp);
      return; // Skip this log entry
    }

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

export { prepareLineChartData, LocalLineChart, LocalPieChart };
