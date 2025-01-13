from dataclasses import dataclass, field
from typing import Any, List

DEFAULT_ARGS = field(default_factory=lambda: [])
DEFAULT_KWARGS = field(default_factory=lambda: {})
DEFAULT_NONE = field(default=None)

@dataclass
class Result:
    task_run_id: str
    result: Any
    std_out: str
    std_err: str
    success: bool

@dataclass
class TaskSetResults:
    task_set_id: str
    results: List[Result]
    success: bool
    overall_status: str

@dataclass
class MapTask:
    name: str
    fun: callable
    args: List[Any] = DEFAULT_ARGS
    kwargs: dict[str, Any] = DEFAULT_KWARGS
    
    def _validate(self):
        if not callable(self.fun):
            raise ValueError("fun must be a callable")
        if not isinstance(self.args, list):
            raise ValueError("args must be a list")
        if not isinstance(self.kwargs, dict):
            raise ValueError("kwargs must be a dict")

@dataclass
class ReduceTask:
    name: str
    fun: callable
    args: List[Any] = DEFAULT_ARGS
    kwargs: dict[str, Any] = DEFAULT_KWARGS
    limit: int = DEFAULT_NONE
    
    def _validate(self):
        if not callable(self.fun):
            raise ValueError("fun must be a callable")
        if not isinstance(self.args, list):
            raise ValueError("args must be a list")
        if not isinstance(self.kwargs, dict):
            raise ValueError("kwargs must be a dict")
        # The first argument of the reduce function must be typed as a list of results
        first_arg_name = list(self.fun.__annotations__.keys())[0]
        provided_type = self.fun.__annotations__.get(first_arg_name)
        if provided_type is None:
            raise ValueError("The first argument of the reduce function must be typed as a list of results")
        if provided_type != List[Result]:
            raise ValueError("The first argument of the reduce function must be typed as a list of results")

@dataclass
class TransformTask:
    name: str
    fun: callable
    args: List[Any] = DEFAULT_ARGS
    kwargs: dict[str, Any] = DEFAULT_KWARGS
    
    def _validate(self):
        if not callable(self.fun):
            raise ValueError("fun must be a callable")
        if not isinstance(self.args, list):
            raise ValueError("args must be a list")
        if not isinstance(self.kwargs, dict):
            raise ValueError("kwargs must be a dict")
        # The first argument of the transform function must be typed as a result
        first_arg_name = list(self.fun.__annotations__.keys())[0]
        provided_type = self.fun.__annotations__.get(first_arg_name)
        if provided_type is None:
            raise ValueError("The first argument of the transform function must be typed as a result")
        if provided_type != Result:
            raise ValueError("The first argument of the transform function must be typed as a result")

@dataclass
class FanoutTask:
    name: str
    fun: callable
    fanout_count: int
    args: List[Any] = DEFAULT_ARGS
    kwargs: dict[str, Any] = DEFAULT_KWARGS
    
    def _validate(self):
        if not callable(self.fun):
            raise ValueError("fun must be a callable")
        if not isinstance(self.args, list):
            raise ValueError("args must be a list")
        if not isinstance(self.kwargs, dict):
            raise ValueError("kwargs must be a dict")
        # The first argument of the fanout function must be typed as a result
        first_arg_name = list(self.fun.__annotations__.keys())[0]
        provided_type = self.fun.__annotations__.get(first_arg_name)
        if provided_type is None:
            raise ValueError("The first argument of the fanout function must be typed as a result")
        if provided_type != Result:
            raise ValueError("The first argument of the fanout function must be typed as a result")
    
# Extended task types 
#  These task types are not native to the GRPC API, but are used internally in the pyharpy library

@dataclass
class BatchMapTask:
    name: str
    batch_size: int
    map_tasks: List[MapTask]

    def _validate(self):
        if not callable(self.fun):
            raise ValueError("fun must be a callable")
        if not isinstance(self.args, list):
            raise ValueError("args must be a list")
        if not isinstance(self.kwargs, dict):
            raise ValueError("kwargs must be a dict")
        
@dataclass
class OneOffClusterTask:
    name: str
    fun: callable
    args: List[Any] = DEFAULT_ARGS
    kwargs: dict[str, Any] = DEFAULT_KWARGS
    
    def _validate(self):
        if not callable(self.fun):
            raise ValueError("fun must be a callable")
        if not isinstance(self.args, list):
            raise ValueError("args must be a list")
        if not isinstance(self.kwargs, dict):
            raise ValueError("kwargs must be a dict")
