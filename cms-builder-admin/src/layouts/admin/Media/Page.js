import {
  Card,
  CardContent,
  CardHeader,
  CardActions,
  Typography,
  Button,
} from "@mui/material";
import { RichTreeView } from "@mui/x-tree-view/RichTreeView";
import { ApiContext } from "../../../context/ApiContext";
import { useEffect, useState, useContext } from "react";
import Grid from "@mui/material/Grid2";

import FilePreview from "./_components/FilePreview";
import { useDialogs } from "../../../context/DialogContext";
import UploadFileForm from "./_components/UploadFileForm";

import { useNotifications } from "../../../context/ToastContext";

export default function MediaPage() {
  const apiService = useContext(ApiContext);
  const dialogs = useDialogs();
  const toast = useNotifications();

  const [fileData, setFileData] = useState([]);
  const [treeItems, setTreeItems] = useState([]);
  const [selectedItem, setSelectedItem] = useState(null);
  const [selectedFile, setSelectedFile] = useState(null);

  const [uploadFormData, setUploadFormData] = useState({
    name: "",
    file: null,
  });

  const [saving, setSaving] = useState(false);

  // Handles upload once save button is clicked
  useEffect(() => {
    if (!saving) return;
    const uploadFile = async () => {
      try {
        const response = await apiService.postFile(uploadFormData.file);
        console.log("File uploaded:", response);
        setSaving(false);
        dialogs.close();
      } catch (error) {
        toast.show("Error uploading file", "error");
        setSaving(false);
      }
    };

    uploadFile();
  }, [saving]);

  const onUploadFile = (data) => {
    setUploadFormData(data);
  };

  useEffect(() => {
    let fn = async () => {
      try {
        const response = await apiService.list("files", 1, 100);
        const files = response.data;
        setFileData(files);
      } catch (error) {
        toast.show(error.message, "error");
      }
    };

    fn();
  }, []);

  useEffect(() => {
    const buildTree = (files) => {
      const root = { id: "root", label: "Media", children: [] };
      const nodes = { root };

      // 1. Separate files and folders:
      const folders = [];
      const filesOnly = [];

      for (let file of files) {
        if (file.includes("/")) {
          folders.push(file);
        } else {
          filesOnly.push(file);
        }
      }

      // 2. Sort folders and files alphabetically:
      folders.sort();
      filesOnly.sort();

      // 3. Build the tree
      const allItems = [...folders, ...filesOnly];

      let idIx = 0;

      for (let file of allItems) {
        const pathParts = file.split("/");
        let currentNode = root;

        for (let i = 0; i < pathParts.length; i++) {
          const part = pathParts[i] || "root";
          const currentPath = pathParts.slice(0, i + 1).join("/");

          if (!nodes[currentPath]) {
            const newNode = {
              id: `${idIx}`, // string id
              path: file,
              label: part,
              children: [],
            };
            nodes[currentPath] = newNode;
            currentNode.children.push(newNode);
          }
          currentNode = nodes[currentPath];
          idIx++;
        }
      }

      // 4. Sort children of each node recursively:
      const sortTree = (node) => {
        if (node.children) {
          node.children.sort((a, b) => a.label.localeCompare(b.label));
          node.children.forEach(sortTree); // Recursively sort children
        }
      };

      sortTree(root); // Start from the root

      return root.children;
    };

    const treeData = [];
    for (let file of fileData) {
      treeData.push(file.path);
    }
    const tree = buildTree(treeData);

    setTreeItems(tree);
  }, [fileData]);

  const find = (tree, id) => {
    for (let i = 0; i < tree.length; i++) {
      if (tree[i].id === id) {
        return tree[i];
      }
      if (tree[i].children && tree[i].children.length > 0) {
        let found = find(tree[i].children, id);
        if (found) {
          return found;
        }
      }
    }
    return null;
  };

  const handleSelection = (e, id, selected) => {
    if (selected) {
      let item = find(treeItems, id);

      if (item.children.length > 0) {
        setSelectedItem(null);
        return;
      }
      setSelectedItem(item);
    }
  };

  const handleUpload = () => {
    const title = "Cargar archivo";
    const content = (
      <UploadFileForm data={uploadFormData} setFormData={onUploadFile} />
    ); // Create content here
    const actions = [
      {
        label: "Guardar",
        onClick: () => {
          setSaving(true);
        },
      },
    ];

    dialogs.show({
      title,
      content,
      actions,
    });
  };

  useEffect(() => {
    if (!selectedItem) {
      return;
    }

    fileData.forEach((file) => {
      if (file.path === selectedItem.path) {
        setSelectedFile(file);
      }
    });
  }, [selectedItem]);

  return (
    <Grid container padding={1} spacing={2}>
      <Grid size={{ xs: 12, sm: 6, md: 5 }}>
        <Card>
          <CardHeader title="Biblioteca de medios" />
          <CardContent
            sx={{
              display: "flex",
              flexDirection: "column",
              gap: 2,
            }}
          >
            <Typography variant="body2">
              Selecciona un archivo para ver su información.
            </Typography>
            {treeItems.length > 0 ? (
              <RichTreeView
                items={treeItems}
                onItemSelectionToggle={handleSelection}
              />
            ) : (
              <Typography variant="body2">No tienes archivos...</Typography>
            )}
          </CardContent>
          <CardActions sx={{ display: "flex", justifyContent: "flex-end" }}>
            <Button size="small" color="primary" onClick={handleUpload}>
              Upload
            </Button>
          </CardActions>
        </Card>
      </Grid>
      <Grid size="grow">
        <FilePreview file={selectedFile} />
      </Grid>
    </Grid>
  );
}
