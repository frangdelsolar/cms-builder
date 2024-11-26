#!/bin/bash


# Get the absolute path to the .gitignore file
gitignore_file=.gitignore

# Check if the .gitignore file exists
if [ -f "$gitignore_file" ]; then
    # Remove the .gitignore file
    rm "$gitignore_file"
    echo "Gitignore file deleted."
else
    echo "Gitignore file not found."
fi

# read every line in the .gitignore file
while IFS= read -r line; do

    # folder in .gitignore is *folderName* should become folderName

    folderName=$(echo "$line" | sed 's/*//g')

    echo "$folderName"
done < .gitignore

echo "Cleaning completed."