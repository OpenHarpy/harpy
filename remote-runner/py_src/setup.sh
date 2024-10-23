#!/bin/bash
# If the venv folder is not present, create it
cd "$(dirname "$0")" # Change to the directory of this script
if [ ! -d "venv" ]; then
    python3 -m venv venv 
    # Activate the virtual environment
    source venv/bin/activate
    pip install -r requirements.txt
    # Find all the whl files in this directory and install them
    for file in $(ls *.whl); do
        echo "Installing $file"
        pip install $file
    done
fi