import axios from "axios";

const apiService = ({ token, apiBaseUrl }) => {
  const withAuth = (config) => {
    if (token) {
      config.headers = {
        ...config.headers,
        Authorization: `Bearer ${token}`,
      };
    }

    return config;
  };
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
      headers: {},
      "Content-Type": contentType,
      data: body,
    });

    try {
      const response = await axios(config);
      return response.data;
    } catch (error) {
      console.error("Error executing API call:", error);
      throw error.response.data;
    }
  };

  const getEntities = () => {
    return executeApiCall({
      method: "GET",
      relativePath: "api",
    });
  };

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

  const postFile = (file) => {
    let path = `private/api/files/new`;

    const formData = new FormData(); // Create a FormData object
    formData.append("file", file);

    return executeApiCall({
      method: "POST",
      relativePath: path,
      body: formData,
      contentType: "multipart/form-data",
    });
  };

  const downloadFile = (fileId) => {
    let path = `private/api/files/${fileId}/download`;

    return executeApiCall({
      method: "GET",
      relativePath: path,
    });
  };

  const getEndpoints = async (entity) => {
    try {
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
    } catch (error) {
      throw error;
    }
  };

  const schema = (entity) => {
    return executeApiCall({
      method: "GET",
      relativePath: `api/${entity}/schema`,
    });
  };

  const post = async (entity, body) => {
    try {
      const response = await executeApiCall({
        method: "POST",
        relativePath: `private/api/${entity}/new`,
        body,
      });
      return response.data;
    } catch (error) {
      throw error;
    }
  };

  const put = async (entity, instance, body) => {
    try {
      const response = await executeApiCall({
        method: "PUT",
        relativePath: `private/api/${entity}/${instance.ID}/update`,
        body,
      });
      return response;
    } catch (error) {
      throw error;
    }
  };

  const destroy = async (resourceName, resourceId) => {
    try {
      return await executeApiCall({
        method: "DELETE",
        relativePath: `private/api/${resourceName}/${resourceId}/delete`,
      });
    } catch (error) {
      throw error;
    }
  };

  // order: fieldName or -fieldName
  const list = async (
    entity,
    page = 1,
    limit = 10,
    order = "",
    query = null
  ) => {
    let path = `private/api/${entity}`;
    const params = new URLSearchParams();
    params.append("page", page);
    params.append("limit", limit);

    if (order) {
      params.append("order", order);
    }
    if (query) {
      for (const [key, value] of Object.entries(query)) {
        if (Array.isArray(value)) {
          params.append(key, value.join(",")); // Handle array values
        } else {
          params.append(key, value);
        }
      }
    }
    try {
      const response = await executeApiCall({
        method: "GET",
        relativePath: `${path}?${params.toString()}`,
      });

      return response;
    } catch (error) {
      throw error;
    }
  };

  const getTimelineForResource = async (
    resourceId,
    resorceName,
    limit,
    page
  ) => {
    const url = `private/api/timeline?resource_id=${resourceId}&resource_name=${resorceName}&limit=${limit}&page=${page}&order=id`;

    return executeApiCall({
      method: "GET",
      relativePath: url,
    });
  };

  const apiUrl = () => {
    return apiBaseUrl;
  };

  return {
    apiUrl,
    getEntities,
    downloadFile,
    getEndpoints,
    getTimelineForResource,
    getRequestLogEntries,
    getRequestStats,
    postFile,
    schema,
    post,
    put,
    destroy,
    list,
  };
};

export default apiService;
