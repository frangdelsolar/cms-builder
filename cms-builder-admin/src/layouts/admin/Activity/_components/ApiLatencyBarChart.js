import React, { useMemo } from "react";
import { BarChart } from "@mui/x-charts/BarChart";
import { Card, CardHeader, CardContent } from "@mui/material";

const ApiLatencyBarChart = ({ data }) => {
  // Process the data to calculate average duration for each API path
  const dataset = useMemo(() => {
    if (!data || !Array.isArray(data)) return [];

    // Aggregate total duration and count for each API path
    const pathStatsMap = data.reduce((acc, request) => {
      const { path, duration } = request;
      if (!acc[path]) {
        acc[path] = { totalDuration: 0, count: 0 };
      }
      acc[path].totalDuration += duration / 1000; // Convert milliseconds to seconds
      acc[path].count += 1;
      return acc;
    }, {});

    // Calculate average duration for each API path
    const pathAverages = Object.entries(pathStatsMap).map(([path, stats]) => ({
      path,
      averageDuration: stats.totalDuration / stats.count, // Average in seconds
    }));

    // Sort by average duration in descending order and take the top 10
    return pathAverages
      .sort((a, b) => b.averageDuration - a.averageDuration)
      .slice(0, 10); // Limit to top 10
  }, [data]);

  // Chart settings
  const chartSetting = {
    height: 350,
    margin: { left: 100 }, // Adjust margin to fit long API paths
  };

  return (
    <Card>
      <CardHeader title="Top 10 API Paths by Average Latency" />
      <CardContent>
        <BarChart
          dataset={dataset}
          yAxis={[{ scaleType: "band", dataKey: "path" }]} // API paths on the y-axis
          series={[
            {
              dataKey: "averageDuration", // Average duration on the x-axis
              label: "Average Duration (s)",
              valueFormatter: (value) => `${value.toFixed(2)} s`, // Format the value to 2 decimal places
            },
          ]}
          layout="horizontal" // Horizontal bar chart
          {...chartSetting}
        />
      </CardContent>
    </Card>
  );
};

export default ApiLatencyBarChart;
