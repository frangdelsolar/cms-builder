import { createSlice } from "@reduxjs/toolkit";

export const projectSlice = createSlice({
  name: "project",
  initialState: {
    data: {
      apiBaseUrl: process.env.NEXT_PUBLIC_API_BASE_URL,
      googleClientId: process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID,
      googleClientSecret: process.env.NEXT_PUBLIC_GOOGLE_CLIENT_SECRET,
      firebaseApiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
      adminEmail: process.env.NEXT_PUBLIC_ADMIN_EMAIL,
      adminPassword: process.env.NEXT_PUBLIC_ADMIN_PASSWORD,
    },
  },
  reducers: {
    setProjectData: (state, action) => {
      state.data = action.payload;
    },
    setApiBaseUrl: (state, action) => {
      state.data.apiBaseUrl = action.payload;
    },
  },
});

export const { setProjectData, setApiBaseUrl } = projectSlice.actions;

export default projectSlice.reducer;

export const selectProjectData = (state) => state.project.data;
