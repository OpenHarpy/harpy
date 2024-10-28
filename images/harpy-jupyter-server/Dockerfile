FROM jupyter/base-notebook:python-3.10

# Get the requirements and whl files
COPY dist/* /tmp/whl_files/
COPY remote-runner/py_src/requirements.txt /tmp/requirements.txt

# Install the requirements
RUN pip install -r /tmp/requirements.txt

# Install the whl files
RUN for file in /tmp/whl_files/*.whl; do \
        echo "Installing $file"; \
        pip install $file; \
    done

# Start the notebook
CMD ["start-notebook.sh", "--NotebookApp.token=''", "--NotebookApp.password=''"]