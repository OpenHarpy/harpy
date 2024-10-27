import cloudpickle
import argparse

def parse_kwarg(kwarg):
    key, value = kwarg.split('=')
    return key, value

def unpickle_block(block_location):
    with open(block_location, 'rb') as f:
        return cloudpickle.loads(f.read())
    

def main():
    parser = argparse.ArgumentParser(description='Run a pickled object.')
    parser.add_argument('--func', type=str, help='The pickled object')
    parser.add_argument('--output', type=str, help='The output file')
    parser.add_argument('--blocks', nargs='*', type=parse_kwarg, help='Key-value pairs')
    parsed_args = parser.parse_args()
    blocks = dict(parsed_args.blocks) if parsed_args.blocks else {}

    # The object is a function, we need to unpickle it
    unpickled_func = unpickle_block(parsed_args.func)
    args = []
    kwargs = {}
    for key, value in blocks.items():
        if '__pos__arg__' in key:
            args.append(unpickle_block(value))
        else:
            kwargs[key] = unpickle_block(value)
    # The object is a class, we need to create an instance of it
    return_object = unpickled_func(*args, **kwargs)
    # The return object needs to be pickled and sent back to the caller
    return_object = cloudpickle.dumps(return_object)
    with open(parsed_args.output, 'wb') as f:
        f.write(return_object)

if __name__ == '__main__':
    main()