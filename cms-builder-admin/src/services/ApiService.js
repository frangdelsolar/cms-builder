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

  const getFiles = () => {
    return executeApiCall({
      method: "GET",
      relativePath: "private/files",
    });
  };

  const downloadFile = (file) => {
    let path = `private/files/download?file=${file}`;

    return executeApiCall({
      method: "GET",
      relativePath: path,
    });
  };

  const getFileInfo = (file) => {
    let path = `private/files/info?file=${file}`;

    return executeApiCall({
      method: "GET",
      relativePath: path,
    });
  };

  const deleteFile = (file) => {
    let path = `private/files/delete?file=${file}`;

    return executeApiCall({
      method: "DELETE",
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
  const list = async (entity, page = 1, limit = 10, order = "") => {
    let path = `private/api/${entity}?page=${page}&limit=${limit}`;
    if (order) {
      path += `&order=${order}`;
    }
    try {
      const response = await executeApiCall({
        method: "GET",
        relativePath: path,
      });

      return response;
    } catch (error) {
      throw error;
    }
  };

  return {
    getEntities,
    getFiles,
    deleteFile,
    downloadFile,
    getFileInfo,
    getEndpoints,
    schema,
    post,
    put,
    destroy,
    list,
  };
};

export default apiService;
