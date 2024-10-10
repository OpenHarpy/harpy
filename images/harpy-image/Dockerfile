FROM harpy-base-go-python:0.0.1 AS harpy-image
# Set the working directory
WORKDIR /app

# Opt folders will be used to store any kind of data that is not considered part of the build or the final binary of the application
# -- This is a good practice to separate the build from the runtime data
RUN mkdir -p /opt/remote-runner/_live_objects
RUN mkdir -p /opt/remote-runner/_python_base
RUN mkdir -p /opt/resource-manager/_harpy_ui
RUN mkdir -p /opt/client-engine/

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

# Build the remote-runner
RUN cd /app/remote-runner/_build && go build -o /app/remote-runner/main
RUN cd /app/resource-manager/_build && go build -o /app/resource-manager/main
RUN cd /app/client-engine/_build && go build -o /app/client-engine/main

# Copy the data that is present in the _build folders to the opt folders (this is the runtime data)
RUN cp -r /app/remote-runner/_build/py_src/* /opt/remote-runner/_python_base/
RUN cp -r /app/remote-runner/_build/whl_files/* /opt/remote-runner/_python_base/
RUN cp -r /app/resource-manager/_build/resource-manager-ui/* /opt/resource-manager/_harpy_ui/

# Now we run setup.sh once to materialize the venv
RUN cd /opt/remote-runner/_python_base && ls && ./setup.sh

# Now we build the default configuration files
#  - remote-runner
RUN cat <<EOF > /app/remote-runner/config.json
{
    "log_level": "info",
    "log_gc": "true",
    "python_installer": "pip3",
    "temporary_binaries": "/opt/remote-runner/_live_objects",
    "command_entrypoint": "/opt/remote-runner/_python_base/entrypoint.sh",
    "enviroment_setup_script": "/opt/remote-runner/_python_base/setup.sh"
}
EOF

#  - resource-manager (config.json, default_env_configs.json, catalog.json)
RUN cat <<EOF > /app/resource-manager/config.json
{
    "log_level": "info",
    "log_gc": "true",
    "node-provider": "local",
    "database-provider": "sqlite",
    "database-uri": "sqlite:///resource-manager.db",
    "default_env_configs": "/app/resource-manager/default_env_configs.json",
    "node_catalog": "/app/resource-manager/catalog.json",
    "local-provider-command": "/app/remote-runner/main local-1 small-4cpu-8gb localhost:50050",
    "ui-static-files": "/opt/resource-manager/_harpy_ui"
}
EOF
RUN cat <<EOF > /app/resource-manager/default_env_configs.json
{
    "harpy.resourceManager.eventLoop.processPoolingIntervalMS" : "1000",
    "harpy.resourceManager.eventLoop.resourceRequestReleasedTimeoutSec" : "60",
    "harpy.resourceManager.eventLoop.requestIdleTimeoutSec" : "120"
}
EOF
RUN cat <<EOF > /app/resource-manager/catalog.json
[
    {
		"NodeType": "small-4cpu-8gb",
		"NumCores": 4,
		"AmountOfMemory": 8,
		"AmountOfStorage": 100,
		"NodeMaxCapacity": 1,
		"NodeWarmpoolSize": 1,
		"NodeIdleTimeout": 60
    }
]
EOF

# - client-engine (config.json)
RUN cat <<EOF > /app/client-engine/config.json
{
    "log_level": "info",
    "log_gc": "true",
    "port": "50051",
    "callback_port": "50052"
}
EOF

# Clean up
RUN rm -rf /app/_build && chmod +x /app/resource-manager/main && chmod +x /app/remote-runner/main

# Set environment variables
ENV REMOTE_RUNNER_CONFIG_FILE /app/remote-runner/config.json
ENV RESOURCE_MANAGER_CONFIG_FILE /app/resource-manager/config.json
ENV CLIENT_ENGINE_CONFIG_FILE /app/client-engine/config.json
ENV COMPONENT UNK

# Dockerfile

# Create a script to run the components
RUN cat <<EOF > /app/run.sh
#!/bin/bash
case "\$COMPONENT" in
    remote-runner)
        /app/remote-runner/main
        ;;
    resource-manager)
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