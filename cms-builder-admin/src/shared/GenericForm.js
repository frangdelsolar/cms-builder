import Form from "@rjsf/mui";
import validator from "@rjsf/validator-ajv8";
import { useEffect, useState } from "react";
import { useAppDispatch, useAppSelector } from "../store/Hooks";
import {
  selectFormData,
  setFormData,
  selectFormErrors,
  selectFormInitialized,
  selectFormSaving,
  setFormInitialized,
  setFormSaving,
  clearForm,
  selectFormSchema,
  selectFormUiSchema,
  selectFormTemplates,
  selectFormWidgets,
  setFormErrors,
} from "../store/FormSlice";

const GenericForm = ({ submitHandler }) => {
  const [formattedErrors, setFormattedErrors] = useState([]);

  const dispatch = useAppDispatch();
  const formData = useAppSelector(selectFormData);
  const formErrors = useAppSelector(selectFormErrors);
  const formSaving = useAppSelector(selectFormSaving);
  const formInitialized = useAppSelector(selectFormInitialized);
  const schema = useAppSelector(selectFormSchema);
  const uiSchema = useAppSelector(selectFormUiSchema);
  const templates = useAppSelector(selectFormTemplates);
  const widgets = useAppSelector(selectFormWidgets);

  useEffect(() => {
    if (!schema) {
      return;
    } else {
      dispatch(setFormInitialized(true));
    }
  }, [schema, uiSchema, templates, widgets, dispatch]);

  useEffect(() => {
    if (formErrors.length > 0) {
      const formatted = formErrors.map((error) => ({
        message: error.message || error,
      }));
      setFormattedErrors(formatted);
    } else {
      setFormattedErrors([]);
    }
  }, [formErrors]);

  useEffect(() => {
    if (formSaving && typeof submitHandler === "function") {
      submitHandler(formData);
      dispatch(setFormSaving(false));
    }
  }, [formSaving, formData, submitHandler]);

  const handleOnChange = (e) => {
    dispatch(setFormData(e.formData));
    dispatch(setFormErrors(e.errors));
  };

  // If no templates are provided, use the default ones
  // SubmitButton is disabled as submit is decoupled
  const getTemplates = () => {
    if (Object.keys(templates).length === 0) {
      return {
        ButtonTemplates: { SubmitButton },
      };
    } else {
      return templates;
    }
  };

  // if (!formInitialized || !schema) {
  //   return <div>Loading form...</div>;
  // }

  return (
    <Form
      uiSchema={uiSchema}
      schema={schema}
      formData={formData}
      onChange={handleOnChange}
      validator={validator}
      templates={getTemplates()}
      widgets={widgets}
      extraErrors={formattedErrors}
    />
  );
};

function SubmitButton() {
  return <></>;
}

export default GenericForm;
