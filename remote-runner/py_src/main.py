import cloudpickle
import argparse
from glob import glob
import io

def kwargs_decode(binary):
    # The binary will always be separated by C0 80 as this will not be a valid UTF-8 character
    split = binary.split(b'\xc0\x80')
    key = split[0].decode('utf-8')
    value = b'\xc0\x80'.join(split[1:])
    buffIO = io.BytesIO(value)
    return key, cloudpickle.load(buffIO)

def input_decode(root, processID):
    # Load the pickled object
    callable_path = f'{root}/{processID}/callable.cloudpickle'
    args = []
    kwargs = {}
    args_glob = glob(f'{root}/{processID}/arg_*.cloudpickle')
    kwargs_glob = glob(f'{root}/{processID}/kwarg_*.cloudpickle')
    for arg in args_glob:
        with open(arg, 'rb') as f:
            args.append(cloudpickle.load(f))
    for kwarg in kwargs_glob:
        with open(kwarg, 'rb') as f:
            key, value = kwargs_decode(f.read())
            kwargs[key] = value
    with open(callable_path, 'rb') as f:
        function = cloudpickle.load(f)
    return args, kwargs, function

def main():
    parser = argparse.ArgumentParser(description='Run a pickled object.')
    parser.add_argument('root', type=str, help='The root directory of the live objects.')
    parser.add_argument('processID', type=str, help='The process ID of the where we will find the binaries.')
    parsed_args = parser.parse_args()
    
    args, kwargs, unpickled_func = input_decode(parsed_args.root, parsed_args.processID)
    # The object is a class, we need to create an instance of it
    return_object = unpickled_func(*args, **kwargs)
    # The return object needs to be pickled and sent back to the caller
    return_object = cloudpickle.dumps(return_object)
    with open(f'{parsed_args.root}/{parsed_args.processID}/return_object.cloudpickle', 'wb') as f:
        f.write(return_object)

if __name__ == '__main__':
    main()
