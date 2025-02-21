import { useEffect, useState } from "react";
import { Button, Box } from "@mui/material";

function UploadFileForm({ setFormData }) {
  const [file, setFile] = useState(null);
  const [buttonLabel, setButtonLabel] = useState("Choose File"); // Initialize to default label

  useEffect(() => {
    setFormData({ file: file });
  }, [file, setFormData]); // Add setFormData to the dependency array

  const handleFileChange = (event) => {
    const selectedFile = event.target.files[0];
    setFile(selectedFile); // Update the file state

    if (selectedFile) {
      setButtonLabel(selectedFile.name);
    } else {
      setButtonLabel("Choose File"); // Reset label if no file selected
    }
  };

  return (
    <Box component="form" noValidate sx={{ mt: 1 }}>
      <Box sx={{ mt: 2 }}>
        <input
          accept="*"
          id="contained-button-file"
          type="file"
          style={{ display: "none" }}
          onChange={handleFileChange}
        />{" "}
        <label htmlFor="contained-button-file">
          <Button variant="contained" component="span">
            {buttonLabel}
          </Button>
        </label>
      </Box>
    </Box>
  );
}

export default UploadFileForm;
