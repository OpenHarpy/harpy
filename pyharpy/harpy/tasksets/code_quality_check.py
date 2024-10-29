import inspect
from harpy.processing.types import (BatchMapTask)
def get_output_type(callable):
    function_signature = inspect.signature(callable)
    return function_signature.return_annotation

def validate_function(callable, function_type, expect_output_type=None):
    # Code quality gates - A. Input and output types are signed
    function_signature = inspect.signature(callable)
    map_output_type = function_signature.return_annotation
    output_signed = map_output_type != inspect._empty
    inputs = function_signature.parameters.values()
    input_signed = all([param.annotation != inspect._empty for param in inputs])
    errors_description = []
    
    if function_type in ["reduce", "transform"] and len(inputs) == 0:
        errors_description.append(function_type.capitalize() + " functions must have at least one argument")
    
    if function_type == "map":
        # Map function need to have typed input and output (there are no restrictions on how the arguments are passed)
        if output_signed and input_signed:
            return []
        else:
            if not output_signed:
                errors_description.append("Map functions must have typed output")
            if not input_signed:
                errors_description.append("Map functions must have typed input")
    elif function_type == "reduce":
        # Reduce function need to have typed output and input 
        # Reduce functions MUST start with VAR_POSITIONAL kind of argument (e.g. *args) and there are no restrictions on how the arguments are passed
        # The first argument MUST be typed according to the EXPECT_OUTPUT_TYPE
        values = list(function_signature.parameters.values())
        check_1 = (output_signed and input_signed)
        check_2 = len(values) > 0 and values[0].kind == inspect.Parameter.VAR_POSITIONAL
        check_3 = values[0].annotation == expect_output_type
        if check_1 and check_2 and check_3:
            return []
        else:
            if not check_1:
                errors_description.append("Reduce functions must have typed input and output")
            if not check_2:
                errors_description.append("Reduce functions must start with VAR_POSITIONAL kind of argument")
            if not check_3:
                errors_description.append("Reduce functions first argument MUST be match the type of the previous function in the chain")
    elif function_type == "transform":
        # Transform function need to have typed input and output (there are no restrictions on how the arguments are passed)
        # They need to have at least one argument and it mustt be POSITIONAL_OR_KEYWORD kind
        # The first argument MUST be typed according to the EXPECT_OUTPUT_TYPE
        values = list(function_signature.parameters.values())
        check_1 = (output_signed and input_signed)
        check_2 = len(values) > 0 and values[0].kind == inspect.Parameter.POSITIONAL_OR_KEYWORD
        check_3 = values[0].annotation == expect_output_type
        if check_1 and check_2 and check_3:
            return []
        else:
            if not check_1:
                errors_description.append("Transform function must have typed input and output")
            if not check_2:
                errors_description.append("Transform function must have at least one argument and it must be POSITIONAL_OR_KEYWORD kind")
            if not check_3:
                errors_description.append("Transform functions first argument MUST be match the type of the previous function in the chain")
    elif function_type == "fanout":
        # Funout function need to have typed input and output (there are no restrictions on how the arguments are passed)
        # They need to have one argument and it must be POSITIONAL_OR_KEYWORD kind
        # The first argument MUST be typed according to the EXPECT_OUTPUT_TYPE
        values = list(function_signature.parameters.values())
        check_1 = (output_signed and input_signed)
        check_2 = len(values) == 1 and values[0].kind == inspect.Parameter.POSITIONAL_OR_KEYWORD
        check_3 = values[0].annotation == expect_output_type
        if check_1 and check_2 and check_3:
            return []
        else:
            if not check_1:
                errors_description.append("Fanout function must have typed input and output")
            if not check_2:
                errors_description.append("Fanout function must have one argument and it must be POSITIONAL_OR_KEYWORD kind")
            if not check_3:
                errors_description.append("Fanout functions first argument MUST be match the type of the previous function in the chain")
    else:
        raise ValueError("Invalid function type")
    return errors_description

def validate_batch_map(batch_maps:BatchMapTask):
    # Code quality gates - B. BatchMapTask can only have maps that have the same function
    errors_description = []
    if len(batch_maps.map_tasks) == 0:
        errors_description.append("BatchMapTask must have at least one map")
        return errors_description
    for map_task in batch_maps.map_tasks:
        if map_task.fun != batch_maps.map_tasks[0].fun:
            errors_description.append("All maps in a BatchMapTask must use the same function")
            break
    error_extend = validate_function(batch_maps.map_tasks[0].fun, "map")
    if len(error_extend) > 0:
        errors_description.extend(error_extend)
    return errors_description
    