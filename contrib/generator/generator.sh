#!/bin/bash

# Check for required arguments
if [ $# -ne 1 ]; then
  echo "Usage: $0 <project-name>"
  exit 1
fi

## VARIABLES ##
project_name="$1"
output_dir="$project_name"
base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
templates_dir="$base_dir/templates"
github_dir="$base_dir/templates/.github"

dockerEnvFile="$templates_dir/.docker.env.template"
developmentEnvFile="$templates_dir/.development.env.template"
testEnvFile="$templates_dir/.test.env.template"
gitignoreFile="$templates_dir/.gitignore.template"

copyFileWithReplacements(){
  src_file="$1"
  dest_file="$2"
  
  # Remove the .template extension from the destination file name
  dest_file="${dest_file%.template}"
  
  cp -f "$src_file" "$dest_file" || { echo "Failed to copy file: $src_file to $dest_file" >&2; exit 1; }
  
  temp_file=$(mktemp)
  sed 's/{{PROJECT_NAME}}/'$project_name'/g' "$dest_file" > "$temp_file"
  mv "$temp_file" "$dest_file"

  temp_file=$(mktemp)
  sed 's/{{projectName}}/'$project_name'/g' "$dest_file" > "$temp_file"
  mv "$temp_file" "$dest_file"

  echo "Copied: $src_file to $dest_file"
}

copyFolderStructure(){
  source_dir="$1"
  dest_dir="$2"

  mkdir -p "$dest_dir" || { echo "Failed to create directory: $dest_dir" >&2; exit 1; }

  # Loop through all items in the source directory, including hidden ones
  for item in "$source_dir/"*; do
    if [ -f "$item" ]; then
      # Copy file with replacements (if needed)
      dest_file="${dest_dir}/${item##*/}"
      # Implement copyFileWithReplacements if needed
      copyFileWithReplacements "$item" "$dest_file"
    elif [ -d "$item" ]; then
      mkdir -p "${dest_dir}/${item##*/}"
      copyFolderStructure "$item" "${dest_dir}/${item##*/}"
    fi
  done
}

updateGitIgnore(){
  echo "*$project_name*" >> $base_dir/.gitignore
}

initGit(){
  cd "$output_dir"
  git init
  cd "$base_dir"
}

# Copy template directory structure
copyFolderStructure "$templates_dir" "$output_dir"

# Hidden stuff to be copied
copyFolderStructure "$github_dir" "$output_dir/.github"
copyFileWithReplacements "$dockerEnvFile" "$output_dir/.docker.env"
copyFileWithReplacements "$developmentEnvFile" "$output_dir/.development.env"
copyFileWithReplacements "$testEnvFile" "$output_dir/.test.env"
copyFileWithReplacements "$gitignoreFile" "$output_dir/.gitignore"

updateGitIgnore
initGit

echo "Project '$project_name' created successfully!"