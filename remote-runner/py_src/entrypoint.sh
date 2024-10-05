#!/bin/bash
current_dir=$(dirname "$0")

# If the venv folder is not present, error out
if [ ! -d "$current_dir/venv" ]; then
    echo "The virtual environment is not present. Please run setup.sh first."
    exit 1
fi

# Activate the virtual environment
echo "Activating the virtual environment"
source $current_dir/venv/bin/activate

# Run the Python script with all the arguments passed to this script
echo "Running the Python script with arguments: $@"
python $current_dir/main.py "$@"