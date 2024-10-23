#!/bin/bash
current_dir=$(dirname "$0")

# The script expects an id as the first argument and a path as the second argument
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <id> <path>"
    exit 1
fi

# If the venv folder is not present, error out
if [ ! -d "$current_dir/venv" ]; then
    echo "The virtual environment is not present. Please run setup.sh first."
    exit 1
fi

# We need to copy the contents of the venv folder to a secondary location (venv-{id}) to isolate the run
# This is because nodes can have multiple runs and we don't want to mix the dependencies
# The venv-{id} folder will be created if it doesn't exist
venv_dir="$current_dir/isolated-$1"
if [ ! -d "$venv_dir" ]; then
    mkdir $venv_dir
    echo "Creating a new virtual environment for isolated-env $1"
    # Copy the contents of the venv folder to the new location
    cp -r $current_dir/venv $venv_dir/venv
    # Copy the entrypoint.sh script to the new location
    cp $2 $venv_dir/entrypoint.sh
    cp $current_dir/main.py $venv_dir/main.py
fi
