#!/bin/bash
current_dir=$(dirname "$0")

# The script expects an id as the first argument
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <id>"
    exit 1
fi

# Define the directory to be deleted
venv_dir="$current_dir/isolated-$1"

# Check if the directory exists
if [ -d "$venv_dir" ]; then
    echo "Removing the virtual environment for isolated-env $1"
    rm -rf "$venv_dir"
    echo "Cleanup completed."
else
    echo "No virtual environment found for isolated-env $1"
fi