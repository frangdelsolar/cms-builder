import { createSlice } from "@reduxjs/toolkit";

export const entitySlice = createSlice({
  name: "entity",
  initialState: {
    selectedEntity: null,
    entities: null,
    schemas: {},
  },
  reducers: {
    setSelectedEntity: (state, action) => {
      state.selectedEntity = action.payload;
    },
    setEntities: (state, action) => {
      state.entities = action.payload;
    },
    setSchema: (state, action) => {
      let key = action.payload.key;
      let schema = action.payload.schema;
      state.schemas[key] = schema;
    },
  },
});

export const { setSelectedEntity, setEntities, setSchema } =
  entitySlice.actions;

export default entitySlice.reducer;

export const selectSelectedEntity = (state) => state.entity.selectedEntity;
export const selectEntities = (state) => state.entity.entities;
export const selectSchemas = (state) => state.entity.schemas;
export const selectEntityTable = (state) => state.entity.table;
