import Grid from "@mui/material/Grid2";
import HistoryEntriesList from "./_components/HistoryEntriesList";
import ErrorMonitor from "./_components/ErrorMonitor";
export default function ActivityPage() {
  return (
    <Grid>
      <Grid container padding={1} spacing={2}>
        <Grid size={{ xs: 12, sm: 3, md: 3 }}>
          <HistoryEntriesList />
        </Grid>
        <Grid size="grow">
          <ErrorMonitor />
        </Grid>
      </Grid>
    </Grid>
  );
}
