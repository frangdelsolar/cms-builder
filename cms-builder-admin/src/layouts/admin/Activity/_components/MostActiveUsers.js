import React from "react";
import { Avatar, Typography, Card, CardContent, Box } from "@mui/material";
import Grid from "@mui/material/Grid2";
import { BarChart } from "@mui/x-charts/BarChart";

const MostActiveUsers = ({ users }) => {
  if (!users || users.length === 0) {
    return (
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Most Active Users
          </Typography>
          <Typography variant="body2" color="textSecondary">
            No users found.
          </Typography>
        </CardContent>
      </Card>
    );
  }
  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          Most Active Users
        </Typography>

        {/* Avatars */}
        <Grid container spacing={2} sx={{ mb: 3 }}>
          {users.map((user) => (
            <Grid item key={user.email} xs={6} sm={4} md={3} lg={2}>
              <Box
                display="flex"
                flexDirection="column"
                alignItems="center"
                textAlign="center"
              >
                <Avatar sx={{ width: 56, height: 56, mb: 1 }}>
                  {user.email?.charAt(0).toUpperCase()}
                </Avatar>
                <Typography
                  variant="body2"
                  sx={{
                    wordBreak: "break-word", // Wrap long text
                    maxWidth: "100px", // Limit the width of the text
                  }}
                >
                  {user.email}
                </Typography>
              </Box>
            </Grid>
          ))}
        </Grid>

        {/* Bar Chart */}
        <BarChart
          dataset={users}
          yAxis={[{ scaleType: "band", dataKey: "email" }]}
          series={[
            {
              dataKey: "count",
              label: "Activity Count",
              valueFormatter: (value) => `${value} activities`,
            },
          ]}
          layout="horizontal"
          height={200}
          margin={{ left: 100 }}
        />
      </CardContent>
    </Card>
  );
};

export default MostActiveUsers;
