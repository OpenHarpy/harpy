from setuptools import setup, find_packages

setup(
    name='sdk',
    version='0.1.0',
    description='A Python client SDK for Project Quack',
    author='Caio',
    author_email='caiopetrellicominato@gmail.com',
    #url='https://github.com/yourusername/project-quack',
    packages=find_packages(),
    install_requires=[
        # List your dependencies here
        'grpcio>=1.66.1',
        'grpcio-tools>=1.66.1',
        'protobuf==5.27.2',
        'cloudpickle>=3.0.0'
    ],
    classifiers=[
        'Programming Language :: Python :: 3',
        'License :: OSI Approved :: MIT License',
        'Operating System :: OS Independent',
    ],
    python_requires='>=3.6',
)