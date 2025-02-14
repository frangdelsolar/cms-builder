import { createSlice } from "@reduxjs/toolkit";

export const formSlice = createSlice({
  name: "form",
  initialState: {
    data: {},
    saving: false,
    initialized: false,
    errors: [],
  },
  reducers: {
    clearForm: (state) => {
      state.data = {};
      state.saving = false;
      state.initialized = false;
      state.errors = [];
    },

    setFormData: (state, action) => {
      state.data = {
        ...state.data,
        ...action.payload,
      };
    },
    setFormErrors: (state, action) => {
      state.errors = action.payload;
    },
    setFormInitialized: (state, action) => {
      state.initialized = action.payload;
    },
    setFormSaving: (state, action) => {
      state.saving = action.payload;
    },
  },
});

export const {
  setFormData,
  setFormErrors,
  setFormInitialized,
  setFormSaving,
  clearForm,
} = formSlice.actions;

export default formSlice.reducer;

export const selectFormData = (state) => state.form.data;
export const selectFormErrors = (state) => state.form.errors;
export const selectFormInitialized = (state) => state.form.initialized;
export const selectFormSaving = (state) => state.form.saving;
