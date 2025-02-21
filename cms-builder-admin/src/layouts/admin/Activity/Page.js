import Grid from "@mui/material/Grid2";
import HistoryEntriesList from "./_components/HistoryEntriesList";
export default function ActivityPage() {
  return (
    <Grid>
      <Grid container padding={1} spacing={2}>
        <Grid size={{ xs: 12, sm: 3, md: 2 }}>Some component</Grid>
        <Grid size="grow">
          <HistoryEntriesList />
        </Grid>
      </Grid>
    </Grid>
  );
}
