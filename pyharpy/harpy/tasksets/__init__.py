from dataclasses import dataclass
import grpc
import cloudpickle
from typing import List

from harpy.primitives import check_variable

from harpy.grpc_ce_protocol.ceprotocol_pb2 import (  
    TaskHandler,
    TaskDefinition,
    TaskSetHandler,
    TaskSetResult,
    TaskResult,
    TaskSetProgressReport,
    
    MapAdder,
    ReduceAdder,
    TransformAdder,
    
    ProgressType
)
from harpy.grpc_ce_protocol.ceprotocol_pb2_grpc import (
    TaskSetStub,
)
from harpy.processing.types import (
    MapTask, ReduceTask, TransformTask, TaskSetResults, Result
)
from harpy.tasksets.code_quality_check import validate_function, get_output_type
from harpy.session.block_read_write_proxy import BlockReadWriteProxy

INVALID_TASKSET_MESSAGE = "TaskSet is not valid, please make sure to create a taskset before using it"
STREAM_CHUNK_SIZE = 1024*1024 # 1MB

class TaskSetDefinitionError(Exception):
    pass

def deserialize_result(result: TaskResult, block_proxy:BlockReadWriteProxy) -> Result:
    output_binary = block_proxy.read_block(result.ObjectReturnBlock)
    stdout_binary = block_proxy.read_block(result.StdoutBlock)
    stderr_binary = block_proxy.read_block(result.StderrBlock)
    # If outputBinary is empty, we return None
    if len(output_binary) == 0:
        py_output = None
    else:
        py_output = cloudpickle.loads(output_binary)
    
    return Result(
        task_run_id=result.TaskRunID,
        result=py_output,
        std_out=stdout_binary.decode(),
        std_err=stderr_binary.decode(),
        success=result.Success
    )

class TaskSet:
    def __init__(self, session, taskSetHandler: TaskSetHandler):
        self._session = session
        self._taskset_stub = TaskSetStub(grpc.insecure_channel(
            f"{self._session.conf.get('sdk.remote.controller.grpcHost')}:{self._session.conf.get('sdk.remote.controller.grpcPort')}"
        ))
        self._taskset_handler = taskSetHandler
        self._last_function_type = None
        self._number_of_nodes_added = 0
    
    def __define_task__(self, task_definition: MapTask | ReduceTask | TransformTask) -> TaskHandler:
        # We start by getting blockIDs for the function and the arguments
        callableBinary = cloudpickle.dumps(task_definition.fun)
        argsBinary = [cloudpickle.dumps(arg) for arg in task_definition.args]
        kwargsBinary = {key: cloudpickle.dumps(value) for key, value in task_definition.kwargs.items()}
        # We then stream all the blocks
        block_proxy = self._session.get_block_read_write_proxy()
        callableBlock = block_proxy.write_block(callableBinary)
        argsBlocks = [block_proxy.write_block(arg) for arg in argsBinary]
        kwargsBlocks = {key: block_proxy.write_block(value) for key, value in kwargsBinary.items()}
        # We then create the task definition chunk
        response:TaskHandler = self._taskset_stub.DefineTask(
            TaskDefinition(
                Name=task_definition.name,
                CallableBlock=callableBlock,
                ArgumentsBlocks=argsBlocks,
                KwargsBlocks=kwargsBlocks
            )
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
            self.__define_task__(map_task)
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
        task_handler = self.__define_task__(reduce_task)
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
        
        task_handler = self.__define_task__(transform_task)
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
            
        print("Getting results")
        response: TaskSetResult = self._taskset_stub.GetTaskSetResults(self._taskset_handler)
        return TaskSetResults(
            task_set_id=taskset_id,
            results=[
                deserialize_result(result, self._session.get_block_read_write_proxy())
                for result in response.TaskResults
            ],
            overall_status = last_taskset_status,
            success = response.OverallSuccess
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
