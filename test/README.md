# Harpy 
These test cases are designed to test the overall functionality of the Harpy library.

These do not test the individual functions of the library but rather test the overall functionality of the library.
For testing the individual functions of the library, you can check the `test` directory in the `harpy` component directory.

Individual tests may not be present in each component directory as they to allow for flexibility during the early stages of development.

## How to run the tests?
To run the tests, you will need to make sure that you have the following items correclty set up:
- Docker and Docker Compose
- Poetry to build the pyharpy package
If this has not been done yet, please refer to the [README.md](../README.md) file for the build instructions.

Once you have the above set up, you can run the tests by running the following command from the root directory of the project:
```bash
make test
```
This will start the tests by using the tests present in this folder.

## Disclaimer
- The tests are not yet complete and may not cover all the functionality of the library.
