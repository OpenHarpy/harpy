from dataclasses import dataclass
import grpc
import cloudpickle
from typing import List

from harpy.primitives import check_variable

from harpy.grpc_ce_protocol.ceprotocol_pb2 import (
    SessionRequest,
    SessionHandler,
    
    TaskHandler,
    TaskDefinitionChunk,
    TaskSetHandler,
    TaskSetResultChunk,
    TaskSetProgressReport,
    
    MapAdder,
    ReduceAdder,
    TransformAdder,
    # Enums
    
    ProgressType
)
from harpy.grpc_ce_protocol.ceprotocol_pb2_grpc import (
    TaskSetStub,
)
from harpy.processing.types import (
    MapTask, ReduceTask, TransformTask, TaskSetResults, Result
)
from harpy.tasksets.code_quality_check import validate_function, get_output_type

INVALID_TASKSET_MESSAGE = "TaskSet is not valid, please make sure to create a taskset before using it"
STREAM_CHUNK_SIZE = 1024*1024 # 1MB

class TaskSetDefinitionError(Exception):
    pass

@dataclass
class TaskDefinitionFull:
    name: str
    fun: bytes
    args: List[bytes]
    kwargs: dict[str, bytes]

@dataclass
class ChunkTaskMapping:
    func_chunk_index: int
    args_chunk_indexes: List[int]
    kwargs_chunk_indexes: dict[str, int]
    def increment(self):
        self.func_chunk_index += 1
        for idx in range(len(self.args_chunk_indexes)):
            self.args_chunk_indexes[idx] += 1
        for key in self.kwargs_chunk_indexes:
            self.kwargs_chunk_indexes[key] += 1

@dataclass
class ResultBinary:
    task_run_id: str
    result: bytes
    std_out: bytes
    std_err: bytes
    success: bool

def make_chunk(data: bytes, chunk_size: int, chunk_index: int) -> bytes:
    # If the chunk index is out of bounds, return an empty byte string
    start = chunk_index * chunk_size
    end = start + chunk_size
    if start >= len(data):
        return b''
    return data[start:end]

def serialize_definition(task_definition: MapTask | ReduceTask | TransformTask) -> TaskDefinitionFull:
    serialized_function = cloudpickle.dumps(task_definition.fun)
    serialized_args = [cloudpickle.dumps(arg) for arg in task_definition.args]
    serialized_kwargs = {key: cloudpickle.dumps(value) for key, value in task_definition.kwargs.items()}
    return TaskDefinitionFull(
        name=task_definition.name,
        fun=serialized_function,
        args=serialized_args,
        kwargs=serialized_kwargs
    )

def deserialize_result(task_result: ResultBinary) -> Result:
    return Result(
        task_run_id = task_result.task_run_id,
        result = cloudpickle.loads(task_result.result) if len(task_result.result) > 0 else None,
        std_out = task_result.std_out.decode(),
        std_err = task_result.std_err.decode(),
        success = task_result.success        
    )

def taskdefinition_generator(task: TaskDefinitionFull):
    task_serialized = serialize_definition(task)
    chunk_mapping = ChunkTaskMapping(
        func_chunk_index=0,
        args_chunk_indexes=[0] * len(task_serialized.args),
        kwargs_chunk_indexes={key: 0 for key in task_serialized.kwargs}
    )
    while True:
        chunk = TaskDefinitionChunk(
            Name=task_serialized.name,
            CallableBinary=make_chunk(task_serialized.fun, STREAM_CHUNK_SIZE, chunk_mapping.func_chunk_index),
            ArgumentsBinary={
                i: make_chunk(arg, STREAM_CHUNK_SIZE, chunk_mapping.args_chunk_indexes[i]) 
                for i, arg in enumerate(task_serialized.args)
            },
            KwargsBinary={
                key: make_chunk(value, STREAM_CHUNK_SIZE, chunk_mapping.kwargs_chunk_indexes[key])
                for key, value in task_serialized.kwargs.items()
            }
        )
        
        # If all the chunks are empty, we are done
        yield chunk
        if (
            (chunk.CallableBinary == b'') and 
            (all(arg == b'' for arg in chunk.ArgumentsBinary.values())) and 
            (all(arg == b'' for arg in chunk.KwargsBinary.values()))
        ):
            break
        chunk_mapping.increment()

class TaskSet:
    def __init__(self, session, taskSetHandler: TaskSetHandler):
        self._session = session
        self._taskset_stub = TaskSetStub(grpc.insecure_channel(
            f"{self._session.conf.get('sdk.remote.controller.grpcHost')}:{self._session.conf.get('sdk.remote.controller.grpcPort')}"
        ))
        self._taskset_handler = taskSetHandler
        self._last_function_type = None
        self._number_of_nodes_added = 0
    
    def __stream_task__(self, task_definition: MapTask | ReduceTask | TransformTask) -> TaskHandler:
        response:TaskHandler = self._taskset_stub.DefineTask(
            taskdefinition_generator(task_definition)
        )
        return response
    
    @check_variable('_taskset_handler', INVALID_TASKSET_MESSAGE)
    def add_maps(self, map_tasks: List[MapTask]) -> 'TaskSet':
        # Map task validation checks
        if self._number_of_nodes_added > 0:
            raise TaskSetDefinitionError("InvalidMapPlacement: Cannot add map tasks after reduce or transform tasks")
        
        output_types = []
        all_errors = []
        for map_index, map_task in enumerate(map_tasks):
            errors = validate_function(map_task.fun, "map")
            output_types.append(get_output_type(map_task.fun))
            if len(errors) > 0:
                error_breakdown = "Map definition at index " + str(map_index) + " failed validation with:"
                for error in errors:
                    error_breakdown += "\n - " + error
                all_errors.append(error_breakdown)
        output_types = list(set(output_types))
        if len(output_types) > 1:
            raise TaskSetDefinitionError("InconsistentMap: All map tasks must have the same output type")
                
        if len(errors) > 0:
            raise TaskSetDefinitionError("InvalidMapFunction: \n" + "\n".join(all_errors))
        self._last_function_type = output_types[0]        
        # Map definition passed all the checks
        task_handlers = [
            self.__stream_task__(map_task)
            for map_task in map_tasks
        ]           
        map_adder = MapAdder(
            taskSetHandler=self._taskset_handler,
            MappersDefinition=task_handlers
        )
        response = self._taskset_stub.AddMap(map_adder)
        if (response.Success):
            self._number_of_nodes_added += 1
            return self
        else:
            raise Exception("Failed to add map tasks")
                    
    @check_variable('_taskset_handler', INVALID_TASKSET_MESSAGE)
    def add_reduce(self, reduce_task: ReduceTask) -> 'TaskSet':
        # Reduce task validation checks
        if self._number_of_nodes_added == 0:
            raise TaskSetDefinitionError("InvalidReducePlacement: Cannot add reduce tasks before map tasks")
        if self._last_function_type is None:
            raise TaskSetDefinitionError("InvalidReducePlacement: Cannot add reduce tasks without previous map tasks")
        
        errors = validate_function(reduce_task.fun, "reduce", self._last_function_type)
        if len(errors) > 0:
            errors = [" - " + error for error in errors]
            raise TaskSetDefinitionError("InvalidReduceFunction:\n" + "\n".join(errors))
        self._last_function_type = get_output_type(reduce_task.fun)
        
        # Reduce definition passed all the checks        
        task_handler = self.__stream_task__(reduce_task)
        reduce_adder = ReduceAdder(
            taskSetHandler=self._taskset_handler,
            ReducerDefinition=task_handler
        )
        response = self._taskset_stub.AddReduce(reduce_adder)
        if (response.Success):
            self._number_of_nodes_added += 1
            return self
        else:
            raise Exception("Failed to add reduce task")
    
    @check_variable('_taskset_handler', INVALID_TASKSET_MESSAGE)
    def add_transform(self, transform_task: TransformTask) -> 'TaskSet':
        # Transform task validation checks
        if self._number_of_nodes_added == 0:
            raise TaskSetDefinitionError("InvalidTransformPlacement: Cannot add transform tasks before map tasks")
        if self._last_function_type is None:
            raise TaskSetDefinitionError("InvalidTransformPlacement: Cannot add transform tasks without previous map tasks")
        
        errors = validate_function(transform_task.fun, "transform", self._last_function_type)
        if len(errors) > 0:
            errors = [" - " + error for error in errors]
            raise TaskSetDefinitionError("InvalidTransformFunction: \n" + "\n".join(errors))
        self._last_function_type = get_output_type(transform_task.fun)
        # Transform definition passed all the checks
        
        task_handler = self.__stream_task__(transform_task)
        transform_adder = TransformAdder(
            taskSetHandler=self._taskset_handler,
            TransformerDefinition=task_handler
        )
        response = self._taskset_stub.AddTransform(transform_adder)
        if (response.Success):
            self._number_of_nodes_added += 1
            return self
        else:
            raise Exception("Failed to add transform task")
    
    @check_variable('_taskset_handler', INVALID_TASKSET_MESSAGE)
    def execute(self) -> TaskSetResults:
        statusSteam: TaskSetProgressReport = self._taskset_stub.Execute(self._taskset_handler)
        # Here we will later add some tracking to the states, for now we just wait for the taskset to finish
        # We will also print the progress as we go
        last_taskset_status = "running"
        taskset_id = self._taskset_handler.taskSetId               
        
        for taskSetStatus in statusSteam:
            if taskSetStatus.ProgressType == ProgressType.TaskSetProgress:
                print(f"TaskSet {taskSetStatus.RelatedID}: {taskSetStatus.ProgressMessage}")
                last_taskset_status = taskSetStatus.StatusMessage
                taskset_id = taskSetStatus.RelatedID
            elif taskSetStatus.ProgressType == ProgressType.TaskGroupProgress:
                print(f"TaskGroup {taskSetStatus.RelatedID} made progress")
            elif taskSetStatus.ProgressType == ProgressType.TaskProgress:
                print(f"Task {taskSetStatus.RelatedID}: {taskSetStatus.StatusMessage}")
            
        # self._taskset_stub.GetTaskSetResults defines a stream of TaskSetResult
        # We need to iterate over the stream to get the results
        print("Getting results")
        responseStream: TaskSetResultChunk = self._taskset_stub.GetTaskSetResults(self._taskset_handler)
        results = {}
        for response in responseStream:
            if response.TaskRunID in results:
                task_result = results[response.TaskRunID]
                task_result.result += response.ObjectReturnBinaryChunk
                task_result.std_out += response.StdoutBinaryChunk
                task_result.std_err += response.StderrBinaryChunk
                task_result.success = response.Success
            else:
                task_result = ResultBinary(
                    task_run_id = response.TaskRunID,
                    result = response.ObjectReturnBinaryChunk,
                    std_out = response.StdoutBinaryChunk,
                    std_err = response.StderrBinaryChunk,
                    success = response.Success
                )
                results[response.TaskRunID] = task_result

        return TaskSetResults(
            task_set_id=taskset_id,
            results=[
                deserialize_result(result)
                for result in results.values()
            ],
            overall_status = last_taskset_status,
            success = True if last_taskset_status == "success" else False
        )
    
    def __dismantle__(self):
        if self._taskset_handler is None:
            return self
        result = self._taskset_stub.Dismantle(self._taskset_handler)
        if (result.Success):
            self._taskset_handler = None
            self._taskset_stub = None
            return self
        else:
            print(result)
            raise Exception("Failed to dismantle taskset")
