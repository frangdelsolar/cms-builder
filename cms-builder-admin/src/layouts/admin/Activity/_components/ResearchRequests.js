import { useState, useEffect } from "react";
import {
  Card,
  CardHeader,
  CardContent,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Button,
} from "@mui/material";
import RequestPreview from "../../Timeline/_components/RequestPreview";

const ResearchRequests = ({ data }) => {
  const [statusCodes, setStatusCodes] = useState([]); // List of unique status codes
  const [selectedStatusCode, setSelectedStatusCode] = useState(""); // Selected status code
  const [apiPaths, setApiPaths] = useState([]); // List of unique API paths for the selected status code
  const [selectedApiPath, setSelectedApiPath] = useState(""); // Selected API path
  const [requestIdentifiers, setRequestIdentifiers] = useState([]); // List of request identifiers for the selected filters
  const [selectedRequestIdentifier, setSelectedRequestIdentifier] =
    useState(""); // Selected request identifier
  const [selectedRequestDetail, setSelectedRequestDetail] = useState(null); // Details of the selected request

  // Extract unique status codes from the data
  useEffect(() => {
    if (data && Array.isArray(data)) {
      const uniqueStatusCodes = [
        ...new Set(data.map((request) => request.status_code)),
      ];
      setStatusCodes(uniqueStatusCodes);
    }
  }, [data]);

  // Update API paths and request identifiers when the selected status code changes
  useEffect(() => {
    if (data && Array.isArray(data)) {
      let filteredRequests = data;

      // Filter by status code if selected
      if (selectedStatusCode) {
        filteredRequests = filteredRequests.filter(
          (request) => request.status_code === selectedStatusCode
        );
      }

      // Extract unique API paths for the filtered requests
      const uniqueApiPaths = [
        ...new Set(filteredRequests.map((request) => request.path)),
      ];
      setApiPaths(uniqueApiPaths);

      // If an API path is selected, ensure it's still valid for the new status code
      if (selectedApiPath && !uniqueApiPaths.includes(selectedApiPath)) {
        setSelectedApiPath(""); // Reset selected API path if it's no longer valid
      }

      // Update request identifiers based on the selected status code and API path
      updateRequestIdentifiers(filteredRequests, selectedApiPath);
    }
  }, [selectedStatusCode, data]);

  // Update request identifiers when the selected API path changes
  useEffect(() => {
    if (data && Array.isArray(data)) {
      let filteredRequests = data;

      // Filter by status code if selected
      if (selectedStatusCode) {
        filteredRequests = filteredRequests.filter(
          (request) => request.status_code === selectedStatusCode
        );
      }

      // Update request identifiers based on the selected status code and API path
      updateRequestIdentifiers(filteredRequests, selectedApiPath);
    }
  }, [selectedApiPath, data]);

  // Helper function to update request identifiers
  const updateRequestIdentifiers = (filteredRequests, apiPath) => {
    if (apiPath) {
      filteredRequests = filteredRequests.filter(
        (request) => request.path === apiPath
      );
    }

    const identifiers = filteredRequests.map(
      (request) => request.request_identifier
    );
    setRequestIdentifiers(identifiers);
    setSelectedRequestIdentifier(""); // Reset selected request identifier
  };

  // Fetch request details when the button is clicked
  const fetchRequestDetail = () => {
    if (selectedRequestIdentifier && data && Array.isArray(data)) {
      const requestDetail = data.find(
        (request) => request.request_identifier === selectedRequestIdentifier
      );
      setSelectedRequestDetail(requestDetail);
    }
  };

  if (!data) {
    return null;
  }

  return (
    <Card>
      <CardHeader title="Researcher" />
      <CardContent>
        {/* Status Code Dropdown */}
        <FormControl
          fullWidth
          variant="outlined"
          style={{ marginBottom: "16px" }}
        >
          <InputLabel>Status Code</InputLabel>
          <Select
            value={selectedStatusCode}
            onChange={(e) => {
              setSelectedStatusCode(e.target.value);
              setSelectedApiPath(""); // Reset API path when status code changes
              setSelectedRequestDetail("");
            }}
            label="Status Code"
          >
            {statusCodes.map((code) => (
              <MenuItem key={code} value={code}>
                {code}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        {/* API Path Dropdown */}
        <FormControl
          fullWidth
          variant="outlined"
          style={{ marginBottom: "16px" }}
        >
          <InputLabel>API Path</InputLabel>
          <Select
            value={selectedApiPath}
            onChange={(e) => {
              setSelectedApiPath(e.target.value);
              setSelectedRequestDetail("");
            }}
            label="API Path"
            disabled={!selectedStatusCode} // Disable if no status code is selected
          >
            {apiPaths.map((path) => (
              <MenuItem key={path} value={path}>
                {path}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        {/* Request Identifier Dropdown */}
        <FormControl
          fullWidth
          variant="outlined"
          style={{ marginBottom: "16px" }}
        >
          <InputLabel>Request Identifier</InputLabel>
          <Select
            value={selectedRequestIdentifier}
            onChange={(e) => {
              setSelectedRequestDetail("");
              setSelectedRequestIdentifier(e.target.value);
            }}
            label="Request Identifier"
            disabled={!selectedStatusCode && !selectedApiPath} // Disable if no filters are selected
          >
            {requestIdentifiers.map((identifier) => (
              <MenuItem key={identifier} value={identifier}>
                {identifier}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        {/* Fetch Details Button */}
        <Button
          variant="contained"
          color="primary"
          onClick={fetchRequestDetail}
          disabled={!selectedRequestIdentifier} // Disable if no request identifier is selected
          style={{ marginBottom: "16px", width: "100%" }}
        >
          Fetch Request Detail
        </Button>

        {/* Display Request Details */}
        {selectedRequestDetail && (
          <RequestPreview
            requestId={selectedRequestDetail.request_identifier}
          />
        )}
      </CardContent>
    </Card>
  );
};

export default ResearchRequests;
