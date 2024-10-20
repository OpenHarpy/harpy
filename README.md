# Harpy 
An open source parallel processing library for your python functions.

Harpy uses a GO backend to run your python functions in parallel. 
It is designed to be simple to use and type restrictive (as much as possible).

This is a WIP project and is not ready for production use yet. (__Read the disclaimer below for more information.__)

## How to use it?
The project is still in development and is not yet available on PyPi or any other package manager.
As of now you need to "build" the project from source in order to start using it.

Efforts are being made to make it available on the docker hub as a docker image and on PyPi as a python package.

Before starting, make sure you have the following installed:
- Docker and Docker Compose
- Poetry to build the pyharpy package

### Building the project
1. Clone the repository
2. Run `make build` to build the docker image and the pyharpy package

After building the project you will find the pyharpy package in the `dist` directory.
The image will be available in your local machine as `harpy:{version}`.

### Running the project

Pyharpy is a client application and it requires that the harpy services are running in the background.
To start the services run `make run-compose` in the root directory of the project. (Chech the `compose_engine` directory for more information)
You can check the `examples` directory for examples on how to use the library.

Once the docker is running a web server will be available at `http://localhost:8080` where you can see the some information about the running services.
Please note that the web server is not yet fully functional and is still in development.

## Disclaimer
- The project as it stands is prone to bugs and is not ready for production use. 
- This is not yet optimized and may not be suitable for large scale applications.
- The project does not yet have a proper testing suite and may not work as expected.
- Currently the application is very "disk hungry" and may consume a lot of disk space.
- The app is prone to memory leaks and may consume a lot of memory.
- Most important of all, if possible, please use a VM for this.

## Contributing
If you want to contribute to the project, please check the `CONTRIBUTING.md` file for more information.
