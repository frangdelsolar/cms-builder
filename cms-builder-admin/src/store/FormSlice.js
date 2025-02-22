import { createSlice } from "@reduxjs/toolkit";

const initialState = {
  data: {},
  schema: {},
  uiSchema: {},
  templates: {},
  widgets: {},
  errors: [],
  file: null,
  saving: false,
  initialized: false,
};

export const formSlice = createSlice({
  name: "form",
  initialState, // Use initialState directly
  reducers: {
    clearForm: (state) => {
      return initialState; // Return a new state object
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

    setFormFile: (state, action) => {
      state.file = action.payload;
    },

    setFormTemplates: (state, action) => {
      state.templates = action.payload;
    },

    setFormWidgets: (state, action) => {
      state.widgets = action.payload;
    },

    setFormUiSchema: (state, action) => {
      state.uiSchema = action.payload;
    },

    setFormSchema: (state, action) => {
      state.schema = action.payload;
    },

    // Example of a reducer that modifies a specific field in the data:
    updateDataField: (state, action) => {
      const { field, value } = action.payload;
      state.data[field] = value;
    },

    // Example of a reducer to add an error:
    addError: (state, action) => {
      state.errors.push(action.payload);
    },

    // Example of a reducer to remove a specific error (by index):
    removeError: (state, action) => {
      const index = action.payload;
      if (index >= 0 && index < state.errors.length) {
        state.errors.splice(index, 1);
      }
    },
  },
});

export const {
  setFormData,
  setFormErrors,
  setFormInitialized,
  setFormSaving,
  setFormFile,
  clearForm,
  setFormTemplates,
  setFormWidgets,
  setFormUiSchema,
  setFormSchema,
  updateDataField, // Added example reducer
  addError, // Added example reducer
  removeError, // Added example reducer
} = formSlice.actions;

export default formSlice.reducer;

// Selectors for all state properties
export const selectFormData = (state) => state.form.data;
export const selectFormErrors = (state) => state.form.errors;
export const selectFormInitialized = (state) => state.form.initialized;
export const selectFormSaving = (state) => state.form.saving;
export const selectFormFile = (state) => state.form.file;
export const selectFormTemplates = (state) => state.form.templates;
export const selectFormWidgets = (state) => state.form.widgets;
export const selectFormUiSchema = (state) => state.form.uiSchema;
export const selectFormSchema = (state) => state.form.schema;
