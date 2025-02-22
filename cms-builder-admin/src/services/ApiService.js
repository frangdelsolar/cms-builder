import axios from "axios";

// Helper function to create the base API service
const createApiService = ({ token, apiBaseUrl }) => {
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

    const config = withAuth({
      method,
      url,
      headers: { "Content-Type": contentType },
      data: body,
    });

    try {
      const response = await axios(config);
      return response.data;
    } catch (error) {
      console.error("Error executing API call:", error);
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
  const getRequestLogEntries = (requestId) => {
    return executeApiCall({
      method: "GET",
      relativePath: `private/api/requests/logs/${requestId}`,
    });
  };

  const getRequestStats = () => {
    return executeApiCall({
      method: "GET",
      relativePath: `private/api/requests/stats`,
    });
  };

  return { getRequestLogEntries, getRequestStats };
};

// File Service
const createFileService = ({ executeApiCall }) => {
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

  const downloadFile = (fileId) => {
    return executeApiCall({
      method: "GET",
      relativePath: `private/api/files/${fileId}/download`,
    });
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
      relativePath: `private/api/timeline?${params.toString()}`,
    });
  };

  return { list, getTimelineForResource };
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

// Main API Service
const apiService = ({ token, apiBaseUrl }) => {
  const { executeApiCall } = createApiService({ token, apiBaseUrl });

  return {
    ...createEntityService({ executeApiCall }),
    ...createRequestService({ executeApiCall }),
    ...createFileService({ executeApiCall }),
    ...createResourceService({ executeApiCall }),
    ...createCrudService({ executeApiCall }),
    apiUrl: () => apiBaseUrl,
  };
};

export default apiService;
