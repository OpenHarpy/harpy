FROM harpy-base-go-python:0.0.1 AS harpy-image
# Set the working directory
WORKDIR /app

# Opt folders will be used to store any kind of data that is not considered part of the build or the final binary of the application
# -- This is a good practice to separate the build from the runtime data
RUN mkdir -p /opt/remote-runner/_blocks
RUN mkdir -p /opt/remote-runner/_python_base
RUN mkdir -p /opt/resource-manager/_harpy_ui
RUN mkdir -p /opt/client-engine/
RUN mkdir -p /opt/configs/

# Set the build directories for the various components
RUN mkdir -p /app/remote-runner/_build
RUN mkdir -p /app/resource-manager/_build
RUN mkdir -p /app/client-engine/_build

# Copy the source code
COPY remote-runner /app/remote-runner/_build
COPY resource-manager /app/resource-manager/_build
COPY client-engine /app/client-engine/_build

# Copy additional whl files
COPY dist/* /app/remote-runner/_build/whl_files/

# Build the components
RUN cd /app/remote-runner/_build && go build -o /app/remote-runner/main
RUN cd /app/resource-manager/_build && go build -o /app/resource-manager/main
RUN cd /app/client-engine/_build && go build -o /app/client-engine/main

# Copy the data that is present in the _build folders to the opt folders (this is the runtime data)
RUN cp -r /app/remote-runner/_build/py_src/* /opt/remote-runner/_python_base/
RUN cp -r /app/remote-runner/_build/whl_files/* /opt/remote-runner/_python_base/
RUN cp -r /app/resource-manager/_build/resource-manager-ui/* /opt/resource-manager/_harpy_ui/

# Now we run setup.sh once to materialize the venv
RUN cd /opt/remote-runner/_python_base && ls && ./setup.sh

# Clean up
RUN rm -rf /app/_build && chmod +x /app/resource-manager/main && chmod +x /app/remote-runner/main

# Set environment variables
ENV RESOURCE_MANAGER_CATALOG_FILE /opt/configs/resource_manager_catalog.json
ENV GLOBAL_CONFIG_FILE /opt/configs/global_env_config.json
ENV COMPONENT UNK

# Dockerfile

# Create a script to run the components
RUN cat <<EOF > /app/run.sh
#!/bin/bash

# Function to check if a file exists
file_exists() {
    if [ ! -f "\$1" ]; then
        echo "File \$1 does not exist"
        exit 1
    fi
}

file_exists "\$GLOBAL_CONFIG_FILE"
case "\$COMPONENT" in
    remote-runner)
        /app/remote-runner/main
        ;;
    resource-manager)
        file_exists "\$RESOURCE_MANAGER_CATALOG_FILE"
        /app/resource-manager/main
        ;;
    client-engine)
        /app/client-engine/main
        ;;
    *)
        echo "Unknown component"
        exit 1
        ;;
esac
EOF

# Set the script as executable
RUN chmod +x /app/run.sh

# Expose the ports
EXPOSE 50050
EXPOSE 50051
EXPOSE 50052
EXPOSE 50053
EXPOSE 8080

# Run the script
CMD ["/app/run.sh"]