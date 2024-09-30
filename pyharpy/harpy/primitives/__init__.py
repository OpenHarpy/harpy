# Singleton primitive
class SingletonMeta(type):
    """
    This is a thread-safe implementation of Singleton.
    """
    _instances = {}

    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:
            instance = super().__call__(*args, **kwargs)
            cls._instances[cls] = instance
        return cls._instances[cls]

# Decorator to check if class_method contains a variable with a specific name not equal to None
def check_variable(variable_name: str, error_message: str):
    def decorator(func):
        def wrapper(*args, **kwargs):
            if not hasattr(args[0], variable_name) or getattr(args[0], variable_name) is None:
                raise ValueError(error_message)
            return func(*args, **kwargs)
        return wrapper
    return decorator