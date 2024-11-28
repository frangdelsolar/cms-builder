#!/bin/bash

# Check for required arguments
if [ $# -ne 1 ]; then
  echo "Usage: $0 <project-name>"
  exit 1
fi

## VARIABLES ##
project_name="$1"

base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
templates_dir="templates"
scripts_dir="$templates_dir/scripts"
github_actions_dir="$templates_dir/.github/workflows"

main_go="$templates_dir/main.go.template"
example_go="$templates_dir/example.go.template"
go_mod="$templates_dir/go.mod.template"
go_sum="$templates_dir/go.sum.template"
config_yaml="$templates_dir/config.yaml.template"
dockerfile="$templates_dir/Dockerfile.template"
dockercompose="$templates_dir/docker-compose.yaml.template"
makefile="$templates_dir/Makefile.template"
gitignore="$templates_dir/.gitignore.template"

find_logs_py="$scripts_dir/find_logs.py.template"
find_todos_py="$scripts_dir/find_todos.py.template"

find_logs_yaml="$github_actions_dir/findLogs.yaml.template"
find_todos_yaml="$github_actions_dir/findTodos.yaml.template"
gofmt_yaml="$github_actions_dir/gofmt.yaml.template"
run_tests_yaml="$github_actions_dir/runTests.yaml.template"

output_dir="$project_name"
output_script_dir="$project_name/scripts"
output_github_actions_dir="$project_name/.github/workflows"


## HELPERS ##
createFolders(){
  mkdir -p "$output_dir" || { echo "Failed to create directory: $output_dir" >&2; exit 1; }
  mkdir -p "$output_dir/cmd" || { echo "Failed to create directory: $output_dir" >&2; exit 1; }
  mkdir -p "$output_script_dir" || { echo "Failed to create directory: $output_script_dir" >&2; exit 1; }
  mkdir -p "$output_github_actions_dir" || { echo "Failed to create directory: $output_github_actions_dir" >&2; exit 1; }
}

updateGitIgnore(){
  echo "*$project_name*" >> $base_dir/.gitignore
}

copyFile(){
  src_file="$1"
  dest_file="$2"
  cp -f "$src_file" "$dest_file" || { echo "Failed to copy file: $src_file to $dest_file" >&2; exit 1; }
  
  temp_file=$(mktemp)
  sed 's/{{PROJECT_NAME}}/'$project_name'/g' "$dest_file" > "$temp_file"
  mv "$temp_file" "$dest_file"

  temp_file=$(mktemp)
  sed 's/{{projectName}}/'$project_name'/g' "$dest_file" > "$temp_file"
  mv "$temp_file" "$dest_file"

  echo "Copied: $src_file to $dest_file"
}


copyFiles(){
  copyFile "$main_go" "$output_dir/cmd/main.go"
  copyFile "$go_mod" "$output_dir/go.mod"
  copyFile "$go_sum" "$output_dir/go.sum"
  copyFile "$example_go" "$output_dir/example.go"
  copyFile "$config_yaml" "$output_dir/config.yaml"
  copyFile "$dockerfile" "$output_dir/Dockerfile"
  copyFile "$dockercompose" "$output_dir/docker-compose.yaml"
  copyFile "$makefile" "$output_dir/Makefile"
  copyFile "$gitignore" "$output_dir/.gitignore"
}

copyScripts(){
  copyFile "$find_logs_py" "$output_script_dir/find_logs.py"
  copyFile "$find_todos_py" "$output_script_dir/find_todos.py"
}

copyGithubActions(){
  copyFile "$find_logs_yaml" "$output_github_actions_dir/findLogs.yaml"
  copyFile "$find_todos_yaml" "$output_github_actions_dir/findTodos.yaml"
  copyFile "$gofmt_yaml" "$output_github_actions_dir/gofmt.yaml"
  copyFile "$run_tests_yaml" "$output_github_actions_dir/runTests.yaml"
}

initGit(){
  cd "$output_dir"
  git init
  cd "$base_dir"
}


## SCRIPT ##
createFolders

updateGitIgnore
copyFiles
copyScripts
copyGithubActions

initGit

echo "Project '$project_name' created successfully!"