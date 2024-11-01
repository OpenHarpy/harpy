from setuptools import setup, find_packages

setup(
    name="pyharpy",
    version="0.1.0",
    description="A Python client for Harpy",
    author="Caio",
    author_email="caiopetrellicominato@gmail.com",
    packages=find_packages(),
    install_requires=[
        # List your dependencies here
        "grpcio>=1.66.1",
        "grpcio-tools>=1.66.1",
        "protobuf==5.27.2",
        "duckdb==1.1.1",
        "dill>=0.3.4",
        "cloudpickle==3.0.0",
        "pandas==2.2.3",
        "ipykernel>=6.29.5",
        "dill>=0.3.8",
        "websockets>=13.0",
        "ipywidgets>=8.1.5",
        "cloudpickle==3.0.0",
        "jinja2>=3.1.4",
        "deltalake==0.20.2",
        "tqdm>=4.66.5",
    ],
    classifiers=[
        "Programming Language :: Python :: 3",
        "License",
        "Operating System :: OS Independent",
    ],
    python_requires=">=3.6",
)