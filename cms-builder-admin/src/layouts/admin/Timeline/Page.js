import ModelList from "../Models/_components/ModelList";
import Grid from "@mui/material/Grid2";
import { useEffect } from "react";

import { setSelectedEntity, selectEntities } from "../../../store/EntitySlice";
import { useAppDispatch, useAppSelector } from "../../../store/Hooks";
import TimelineItemPreview from "./_components/TimelineItemPreview";

export default function TimelinePage() {
  const dispatch = useAppDispatch();
  const entities = useAppSelector(selectEntities);

  useEffect(() => {
    if (!entities) {
      return;
    }

    let hash = window.location.hash;

    if (!hash) {
      hash = entities[0].kebabPluralName;
      window.location.hash = hash;
    }

    let entity = entities.find((e) => e.kebabPluralName === hash.slice(1));
    dispatch(setSelectedEntity(entity));
  }, []);

  return (
    <Grid>
      <Grid container padding={1} spacing={2}>
        <Grid size={{ xs: 12, sm: 3, md: 2 }}>
          <ModelList />
        </Grid>
        <Grid size="grow">
          <TimelineItemPreview />
        </Grid>
      </Grid>
    </Grid>
  );
}
