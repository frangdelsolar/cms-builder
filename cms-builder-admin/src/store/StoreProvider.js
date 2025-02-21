import { Provider } from "react-redux";
import { configureStore } from "@reduxjs/toolkit";
import { entitySlice } from "./EntitySlice";
import { formSlice } from "./FormSlice";

const preloadedState = window.__PRELOADED_STATE__;

const store = configureStore({
  reducer: {
    entity: entitySlice.reducer,
    form: formSlice.reducer,
  },
});

export default function StoreProvider({ children }) {
  return (
    <Provider store={store} serverState={preloadedState}>
      {children}
    </Provider>
  );
}
