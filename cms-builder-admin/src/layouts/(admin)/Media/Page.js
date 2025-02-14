import { Card, CardContent, CardHeader } from "@mui/material";
import { RichTreeView } from "@mui/x-tree-view/RichTreeView";
import { ApiContext } from "../../../context/ApiContext";
import { useEffect, useState, useContext } from "react";
import Grid from "@mui/material/Grid2";

import FilePreview from "./_components/FilePreview";

export default function MediaPage() {
  const apiService = useContext(ApiContext);

  const [treeItems, setTreeItems] = useState([]);
  const [selectedItem, setSelectedItem] = useState(null);

  useEffect(() => {
    const buildTree = (files) => {
      const root = { id: "root", label: "Media", children: [] };
      const nodes = { root };

      // 1. Separate files and folders:
      const folders = [];
      const filesOnly = [];

      for (let file of files) {
        if (file.startsWith("/")) {
          file = file.slice(1);
        }

        if (file.includes("/")) {
          // It's a folder or file within a folder
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
          const part = pathParts[i];
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

    let fn = async () => {
      try {
        const response = await apiService.getFiles();
        const files = response.data;

        const tree = buildTree(files);

        setTreeItems(tree);
      } catch (error) {
        console.error("Error fetching files:", error);
      }
    };

    fn();
  }, [apiService]);

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

  return (
    <Grid container padding={1} spacing={2}>
      <Grid size={{ xs: 12, sm: 6, md: 5 }}>
        <Card>
          <CardHeader title="Biblioteca" />
          <CardContent>
            <RichTreeView
              items={treeItems}
              onItemSelectionToggle={handleSelection}
            />
          </CardContent>
        </Card>
      </Grid>
      <Grid size="grow">
        <FilePreview file={selectedItem} />
      </Grid>
    </Grid>
  );
}
