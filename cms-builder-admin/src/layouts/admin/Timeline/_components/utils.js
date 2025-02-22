const formatChanges = (changes, indent = "") => {
  if (!changes || typeof changes !== "object") return "No changes";

  let formattedChanges = [];
  for (const key in changes) {
    if (changes.hasOwnProperty(key)) {
      const changeValue = changes[key];
      if (Array.isArray(changeValue) && changeValue.length === 2) {
        formattedChanges.push(
          `${indent}${key}: ${JSON.stringify(
            changeValue[0]
          )} -> ${JSON.stringify(changeValue[1])}`
        );
      } else if (typeof changeValue === "object" && changeValue !== null) {
        const nestedFormattedChanges = formatChanges(
          changeValue,
          indent + "  "
        ); // Add indentation
        if (nestedFormattedChanges) {
          formattedChanges.push(`${indent}${key}: `);
          formattedChanges.push(nestedFormattedChanges);
          formattedChanges.push(`${indent}`);
        }
      } else {
        formattedChanges.push(
          `${indent}${key}: ${JSON.stringify(changeValue)}`
        );
      }
    }
  }
  return formattedChanges.join("\n");
};

export { formatChanges };
