import ModelList from "./_components/ModelList";
import ModelTable from "./_components/ModelTable";
import HistoryEntriesList from "./_components/HistoryEntriesList";
import Grid from "@mui/material/Grid2";
import { useEffect } from "react";

import { setSelectedEntity, selectEntities } from "../../../store/EntitySlice";
import { useAppDispatch, useAppSelector } from "../../../store/Hooks";

export default function EntitiesPage() {
  const dispatch = useAppDispatch();
  const entities = useAppSelector(selectEntities);

  useEffect(() => {
    let hash = window.location.hash;
    if (hash && entities) {
      let entity = entities.find((e) => e.kebabPluralName === hash.slice(1));
      dispatch(setSelectedEntity(entity));
    }
  }, []);

  return (
    <Grid>
      <Grid container padding={1} spacing={2}>
        <Grid size={{ xs: 12, sm: 3, md: 2 }}>
          <ModelList />
        </Grid>
        <Grid size="grow">
          <ModelTable />
        </Grid>
        <Grid size={{ xs: 12, sm: 4, md: 3 }}>
          <HistoryEntriesList />
        </Grid>
      </Grid>
    </Grid>
  );
}
