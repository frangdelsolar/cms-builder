import axios from "axios";

// Helper function to create the base API service
const createApiService = ({ token, apiBaseUrl, origin }) => {
  // Helper function to add authorization headers
  const withAuth = (config) => {
    if (token) {
      config.headers = {
        ...config.headers,
        Authorization: `Bearer ${token}`,
      };
    }
    return config;
  };

  // Centralized API call executor
  const executeApiCall = async ({
    method,
    relativePath,
    body = {},
    contentType = "application/json",
  }) => {
    const url = `${apiBaseUrl}/${relativePath}`;

    const headers = {
      "Content-Type": contentType,
    };

    if (origin) {
      headers.Origin = origin;
    }

    const config = withAuth({
      method,
      url,
      headers: headers,
      data: body,
    });

    try {
      const response = await axios(config);
      return response.data;
    } catch (error) {
      throw error.response?.data || error.message;
    }
  };

  return { executeApiCall };
};

// Entity Service
const createEntityService = ({ executeApiCall }) => {
  const getEntities = () => {
    return executeApiCall({
      method: "GET",
      relativePath: "api",
    });
  };

  const getEndpoints = async (entity) => {
    const response = await executeApiCall({
      method: "GET",
      relativePath: "api",
    });
    const entitiesDetails = response.data;
    const validEntity = entitiesDetails.find(
      (item) => item.pluralName === entity
    );

    if (!validEntity) {
      throw new Error("Invalid entity");
    }
    return validEntity;
  };

  const schema = (entity) => {
    return executeApiCall({
      method: "GET",
      relativePath: `api/${entity}/schema`,
    });
  };

  return { getEntities, getEndpoints, schema };
};

// Request Service
const createRequestService = ({ executeApiCall }) => {
  const getRequestLogEntries = (traceId) => {
    return executeApiCall({
      method: "GET",
      relativePath: `private/request-logs/${traceId}`,
    });
  };

  const getRequestStats = () => {
    return executeApiCall({
      method: "GET",
      relativePath: `private/request-logs-stats`,
    });
  };

  return { getRequestLogEntries, getRequestStats };
};

// File Service
const createFileService = ({ executeApiCall, apiBaseUrl, token }) => {
  const postFile = (file) => {
    const formData = new FormData();
    formData.append("file", file);

    return executeApiCall({
      method: "POST",
      relativePath: "private/api/files/new",
      body: formData,
      contentType: "multipart/form-data",
    });
  };

  const downloadFile = async (fileId) => {
    const url = `${apiBaseUrl}/private/api/files/${fileId}/download`;

    const headers = {
      Authorization: `Bearer ${token}`, // Use the token from the service
    };

    try {
      const response = await axios({
        method: "GET",
        url: url,
        headers: headers,
        responseType: "blob", // Ensure the response is treated as a binary blob
      });

      return response; // Return the raw Response object
    } catch (error) {
      console.error("Error downloading file:", error);
      throw error;
    }
  };

  return { postFile, downloadFile };
};
// Resource Service
const createResourceService = ({ executeApiCall }) => {
  const list = async (
    entity,
    page = 1,
    limit = 10,
    order = "",
    query = null
  ) => {
    const params = new URLSearchParams({ page, limit, order });
    if (query) {
      Object.entries(query).forEach(([key, value]) => {
        params.append(key, Array.isArray(value) ? value.join(",") : value);
      });
    }

    return executeApiCall({
      method: "GET",
      relativePath: `private/api/${entity}?${params.toString()}`,
    });
  };

  const getTimelineForResource = async (
    resourceId,
    resourceName,
    limit,
    page
  ) => {
    const params = new URLSearchParams({
      resource_id: resourceId,
      resource_name: resourceName,
      limit,
      page,
      order: "id",
    });
    return executeApiCall({
      method: "GET",
      relativePath: `private/api/database-timeline?${params.toString()}`,
    });
  };

  return { list, getTimelineForResource };
};

const createJobService = ({ executeApiCall }) => {
  const runJob = async (jobName) => {
    const params = new URLSearchParams({ job_definition_name: jobName });

    return executeApiCall({
      method: "POST",
      relativePath: `private/job/run?${params.toString()}`,
    });
  };

  return { runJob };
};

// CRUD Service
const createCrudService = ({ executeApiCall }) => {
  const post = (entity, body) => {
    return executeApiCall({
      method: "POST",
      relativePath: `private/api/${entity}/new`,
      body,
    });
  };

  const put = (entity, instance, body) => {
    return executeApiCall({
      method: "PUT",
      relativePath: `private/api/${entity}/${instance.ID}/update`,
      body,
    });
  };

  const destroy = (resourceName, resourceId) => {
    return executeApiCall({
      method: "DELETE",
      relativePath: `private/api/${resourceName}/${resourceId}/delete`,
    });
  };

  return { post, put, destroy };
};

const healthCheck = ({ executeApiCall }) => {
  return {
    healthCheck: () =>
      executeApiCall({
        method: "GET",
        relativePath: "", // Adjust the path as needed
      }),
  };
};

// Main API Service
const apiService = ({ token, apiBaseUrl, origin }) => {
  const { executeApiCall } = createApiService({ token, apiBaseUrl });

  return {
    ...createEntityService({ executeApiCall }),
    ...createRequestService({ executeApiCall }),
    ...createFileService({ executeApiCall, apiBaseUrl, token }),
    ...createResourceService({ executeApiCall }),
    ...createCrudService({ executeApiCall }),
    ...healthCheck({ executeApiCall }),
    ...createJobService({ executeApiCall }),
    apiUrl: () => apiBaseUrl,
  };
};

export default apiService;
