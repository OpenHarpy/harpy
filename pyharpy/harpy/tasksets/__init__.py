from abc import ABC, abstractmethod
import grpc
import cloudpickle
from typing import List, Any, Optional

from harpy.primitives import check_variable
from harpy.exceptions.user_facing import TaskSetRuntimeError
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
    FanoutAdder,
    OneOffClusterAdder,
    
    ProgressType
)
from harpy.grpc_ce_protocol.ceprotocol_pb2_grpc import (
    TaskSetStub,
)
from harpy.processing.types import (
    MapTask, 
    BatchMapTask,
    FanoutTask,
    ReduceTask, 
    OneOffClusterTask,
    TransformTask, 
    TaskSetResults,
    Result
)
from harpy.tasksets.code_quality_check import validate_function, get_output_type, validate_batch_map
from harpy.session.block_read_write_proxy import BlockReadWriteProxy
from harpy.tasksets.progress_displays import get_progress_viewer

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

def get_block_for_object(block_proxy:BlockReadWriteProxy, block_hash_map:dict, obj:Any) -> str:
    if id(obj) in block_hash_map:
        return block_hash_map[id(obj)]
    obj_binary = cloudpickle.dumps(obj)
    block_hash = block_proxy.write_block(obj_binary)
    block_hash_map[id(obj)] = block_hash
    return block_hash

class TaskSetClient:
    def __init__(self, session, options={}):
        self.session = session
        self.taskset_stub = TaskSetStub(grpc.insecure_channel(
            f"{self.session.conf.get('harpy.sdk.remote.controller.grpcHost')}:{self.session.conf.get('harpy.sdk.remote.controller.grpcPort')}"
        ))
        self.taskset_handler = self.session.get_raw_taskset_handler(raw_tasksets_close_callbacks=self.dismantle, options=options)
        self.block_hash_map = {}
    
    def __get_block_for_object__(self, obj:Any) -> str:
        return get_block_for_object(self.session.get_block_read_write_proxy(), self.block_hash_map, obj)
    
    @check_variable('taskset_handler', INVALID_TASKSET_MESSAGE)
    def define_task(self, task_definition: MapTask | ReduceTask | TransformTask) -> TaskHandler:
        # We start by getting blockIDs for the function and the arguments        
        callableBlock = self.__get_block_for_object__(task_definition.fun)
        argsBlocks = [self.__get_block_for_object__(arg) for arg in task_definition.args]
        kwargsBlocks = {key: self.__get_block_for_object__(value) for key, value in task_definition.kwargs.items()}
        # We then create the task definition chunk
        response:TaskHandler = self.taskset_stub.DefineTask(
            TaskDefinition(
                Name=task_definition.name,
                CallableBlock=callableBlock,
                ArgumentsBlocks=argsBlocks,
                KwargsBlocks=kwargsBlocks
            )
        )                   
        return response
    
    def __track_progress__(self, status_stream, pview):
        pview.set_context_info("Running...")
        last_taskset_status = "running"
        current_taskgroup = ""
        taskset_id = None
    
        for taskSetStatus in status_stream:
            details = taskSetStatus.ProgressDetails
            if taskSetStatus.ProgressType == ProgressType.TaskSetProgress:
                last_taskset_status = taskSetStatus.StatusMessage
                taskset_id = taskSetStatus.RelatedID
            elif taskSetStatus.ProgressType == ProgressType.TaskGroupProgress:
                current_taskgroup = details['taskgroup_name']
            elif taskSetStatus.ProgressType == ProgressType.TaskProgress:
                if current_taskgroup != "":
                    pview.add_step(current_taskgroup, taskSetStatus.RelatedID)
                    if taskSetStatus.StatusMessage == "running":
                        pview.task_running(current_taskgroup, taskSetStatus.RelatedID)
                    elif taskSetStatus.StatusMessage == "done":
                        taskrun_success = details.get('taskrun_success')
                        if taskrun_success == 'true':
                            pview.task_success(current_taskgroup, taskSetStatus.RelatedID)
                        elif taskrun_success == 'false':
                            pview.task_fail(current_taskgroup, taskSetStatus.RelatedID)
                    elif taskSetStatus.StatusMessage == "panic":
                        pview.task_fail(current_taskgroup, taskSetStatus.RelatedID)
    
        return last_taskset_status, taskset_id
    
    @check_variable('taskset_handler', INVALID_TASKSET_MESSAGE)
    def execute(self, collect=False) -> TaskSetResults:
        status_stream: TaskSetProgressReport = self.taskset_stub.Execute(self.taskset_handler)
        
        pview = get_progress_viewer()
        last_taskset_status, taskset_id = self.__track_progress__(status_stream, pview)
        
        response: TaskSetResult = self.taskset_stub.GetTaskSetResults(self.taskset_handler)
        results = None
        
        if collect:
            pview.set_context_info("Collecting results...", context_working=True)
            results = [
                deserialize_result(result, self.session.get_block_read_write_proxy())
                for result in response.TaskResults
            ]
        pview.set_overall_status(response.OverallSuccess)
        pview.set_context_info("Done")
        return TaskSetResults(
            task_set_id=taskset_id,
            results=results,
            overall_status=last_taskset_status,
            success=response.OverallSuccess
        )
    
    def dismantle(self):
        if self.taskset_handler is not None:
            result = self.taskset_stub.Dismantle(self.taskset_handler)
            if (result.Success):
                self.taskset_handler = None
                self.taskset_stub = None
                return self
            else:
                print(result)
                raise Exception("Failed to dismantle taskset")
    
    def __check_response__(self, response):
        if not response.Success:
            raise Exception(response.ErrorMessage)
        return response
    
    @check_variable('taskset_handler', INVALID_TASKSET_MESSAGE)
    def add_map(self, map_definitions: List[MapTask]):
        task_handlers = [
            self.define_task(map_task)
            for map_task in map_definitions
        ]
        map_adder = MapAdder(
            taskSetHandler=self.taskset_handler,
            MappersDefinition=task_handlers
        )
        self.__check_response__(self.taskset_stub.AddMap(map_adder))

    @check_variable('taskset_handler', INVALID_TASKSET_MESSAGE)
    def add_reduce(self, reduce_task: ReduceTask):
        task_handler = self.define_task(reduce_task)
        limit = -1
        if reduce_task.limit is not None:
            limit = reduce_task.limit
        reduce_adder = ReduceAdder(
            taskSetHandler=self.taskset_handler,
            ReducerDefinition=task_handler,
            Limit=limit
        )
        self.__check_response__(self.taskset_stub.AddReduce(reduce_adder))
    
    @check_variable('taskset_handler', INVALID_TASKSET_MESSAGE)
    def add_transform(self, transform_task: TransformTask):
        task_handler = self.define_task(transform_task)
        transform_adder = TransformAdder(
            taskSetHandler=self.taskset_handler,
            TransformerDefinition=task_handler
        )
        self.__check_response__(self.taskset_stub.AddTransform(transform_adder))
    
    @check_variable('taskset_handler', INVALID_TASKSET_MESSAGE)
    def add_fanout(self, fanout_task: FanoutTask):
        task_handler = self.define_task(fanout_task)
        fanout_adder = FanoutAdder(
            taskSetHandler=self.taskset_handler,
            FanoutDefinition=task_handler,
            FanoutCount=fanout_task.fanout_count
        )
        self.__check_response__(self.taskset_stub.AddFanout(fanout_adder))        

    @check_variable('taskset_handler', INVALID_TASKSET_MESSAGE)
    def add_oneoff_cluster(self, oneoff_cluster_task: OneOffClusterTask):
        task_handler = self.define_task(oneoff_cluster_task)
        oneoff_cluster_adder = OneOffClusterAdder(
            taskSetHandler=self.taskset_handler,
            OneOffClusterDefinition=task_handler
        )
        self.__check_response__(self.taskset_stub.AddOneOffCluster(oneoff_cluster_adder))

class NodeLazy(ABC):
    @abstractmethod
    def add_to_taskset(self, taskset_client: TaskSetClient):
        pass
    @abstractmethod
    def __repr__(self):
        pass

class NodeLazyMap(NodeLazy):
    def __init__(self, map_definitions: list[MapTask]):
        self.node_type = 'Map'
        self.definitions = map_definitions

    def __repr__(self):
        unique_names = list(set([definition.name for definition in self.definitions]))
        names_str = ", ".join(unique_names)
        return f"{self.node_type} ({names_str}) [total_functions={len(self.definitions)}]"

    def add_to_taskset(self, taskset_client: TaskSetClient):
        taskset_client.add_map(self.definitions)

class NodeLazyReduce(NodeLazy):
    def __init__(self, reduce_definition: ReduceTask):
        self.node_type = 'Reduce'
        self.definition = reduce_definition
    
    def __repr__(self):
        return f"{self.node_type} ({self.definition.name})"
    
    def add_to_taskset(self, taskset_client: TaskSetClient):
        taskset_client.add_reduce(self.definition)

class NodeLazyTransform(NodeLazy):
    def __init__(self, transform_definition: TransformTask):
        self.node_type = 'Transform'
        self.definition = transform_definition
    
    def __repr__(self):
        return f"{self.node_type} ({self.definition.name})"
    
    def add_to_taskset(self, taskset_client: TaskSetClient,):
        taskset_client.add_transform(self.definition)
    
class NodeLazyFanout(NodeLazy):
    def __init__(self, fanout_definition: FanoutTask):
        self.node_type = 'Fanout'
        self.definition = fanout_definition
    
    def __repr__(self):
        return f"{self.node_type} ({self.definition.name}) [fanout_count={self.definition.fanout_count}]"
    
    def add_to_taskset(self, taskset_client: TaskSetClient,):
        taskset_client.add_fanout(self.definition)

class NodeLazyBatchMap(NodeLazy):
    def __init__(self, batchmap_definition: BatchMapTask):
        self.node_type = 'BatchMap'
        self.definition = batchmap_definition
        
        # We will wrap the batch map task into chunked map tasks based on the batch size
        passed_map_tasks_size = len(self.definition.map_tasks)
        tasks_per_chunk = self.definition.batch_size        
        # Create the chunked map tasks
        chunked_map_tasks = [
            self.definition.map_tasks[i:i + tasks_per_chunk]
            for i in range(0, passed_map_tasks_size, tasks_per_chunk)
        ]
        # We will create a map task for each chunk
        def map_wrap(chunk: list[MapTask]) -> list[Any]:
            return [task.fun(*task.args, **task.kwargs) for task in chunk]
        
        self.definitions = [
            MapTask(
                name = f'chunked_map_{i}',
                fun=map_wrap,
                kwargs={"chunk": chunk}
            )
            for i, chunk in enumerate(chunked_map_tasks)
        ]
    
    def __repr__(self):
        return f"{self.node_type} ({self.definition.name}) with {len(self.definitions)} definitions"
    
    def add_to_taskset(self, taskset_client: TaskSetClient,):
        taskset_client.add_map(self.definitions)

class NodeLazyOneOffCluster(NodeLazy):
    def __init__(self, oneoff_cluster_definition: OneOffClusterTask):
        self.node_type = 'OneOffCluster'
        self.definition = oneoff_cluster_definition
    
    def __repr__(self):
        return f"{self.node_type} ({self.definition.name})"
    
    def add_to_taskset(self, taskset_client: TaskSetClient,):
        taskset_client.add_oneoff_cluster(self.definition)

class TaskSet:
    def __init__(self, session, options={'harpy.taskset.name': 'pyharpy-unamed-taskset'}):
        self._session = session
        self._taskset_client = None
        self._last_function_type = None
        self._block_hash_map = {}
        self._lazy_nodes:list[NodeLazy] = []
        self._options = options
    
    def add_maps(self, map_tasks: List[MapTask]) -> 'TaskSet':
        # Map task validation checks
        if len(self._lazy_nodes) > 0:
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
        self._lazy_nodes.append(NodeLazyMap(map_tasks,))
        return self

    def add_reduce(self, reduce_task: ReduceTask) -> 'TaskSet':
        # Reduce task validation checks
        if len(self._lazy_nodes) == 0:
            raise TaskSetDefinitionError("InvalidReducePlacement: Cannot add reduce tasks before map tasks")
        if self._last_function_type is None:
            raise TaskSetDefinitionError("InvalidReducePlacement: Cannot add reduce with tasks that have no output type")
        
        errors = validate_function(reduce_task.fun, "reduce", self._last_function_type)
        if len(errors) > 0:
            errors = [" - " + error for error in errors]
            raise TaskSetDefinitionError("InvalidReduceFunction:\n" + "\n".join(errors))
        if reduce_task.limit is not None:
            if reduce_task.limit <= 0:
                raise TaskSetDefinitionError("InvalidReduceLimit: Reduce limit must be a positive integer and greater than 0")
        self._last_function_type = get_output_type(reduce_task.fun)
        # Reduce definition passed all the checks
        self._lazy_nodes.append(NodeLazyReduce(reduce_task,))
        return self
    
    def add_transform(self, transform_task: TransformTask) -> 'TaskSet':
        # Transform task validation checks
        if len(self._lazy_nodes) == 0:
            raise TaskSetDefinitionError("InvalidTransformPlacement: Cannot add transform tasks before map tasks")
        if self._last_function_type is None:
            raise TaskSetDefinitionError("InvalidTransformPlacement: Cannot transform with tasks that have no output type")
        
        errors = validate_function(transform_task.fun, "transform", self._last_function_type)
        if len(errors) > 0:
            errors = [" - " + error for error in errors]
            raise TaskSetDefinitionError("InvalidTransformFunction: \n" + "\n".join(errors))
        self._last_function_type = get_output_type(transform_task.fun)
        # Transform definition passed all the checks
        self._lazy_nodes.append(NodeLazyTransform(transform_task,))
        return self
    
    def add_fanout(self, fanout_task: FanoutTask) -> 'TaskSet':
        # Fanout task validation checks
        if len(self._lazy_nodes) == 0:
            raise TaskSetDefinitionError("InvalidFanoutPlacement: Cannot add fanout tasks before map tasks")
        if self._last_function_type is None:
            raise TaskSetDefinitionError("InvalidFanoutPlacement: Cannot add fanout tasks without previous map tasks")
        
        errors = validate_function(fanout_task.fun, "fanout", self._last_function_type)
        if len(errors) > 0:
            errors = [" - " + error for error in errors]
            raise TaskSetDefinitionError("InvalidFanoutFunction: \n" + "\n".join(errors)
        )
        self._last_function_type = get_output_type(fanout_task.fun)
        # Fanout definition passed all the checks
        self._lazy_nodes.append(NodeLazyFanout(fanout_task,))
        return self 
    
    def add_oneoff_cluster(self, oneoff_cluster_task: OneOffClusterTask) -> 'TaskSet':
        # OneOffCluster task validation checks
        if len(self._lazy_nodes) > 0:
            raise TaskSetDefinitionError("InvalidOneOffClusterPlacement: Cannot add oneoff cluster tasks after reduce or transform tasks")
        
        errors = validate_function(oneoff_cluster_task.fun, "map") # We use map validation for oneoff cluster tasks as they are similar
        if len(errors) > 0:
            errors = [" - " + error for error in errors]
            raise TaskSetDefinitionError("InvalidOneOffClusterFunction: \n" + "\n".join(errors))
        self._last_function_type = get_output_type(oneoff_cluster_task.fun)
        # OneOffCluster definition passed all the checks
        self._lazy_nodes.append(NodeLazyOneOffCluster(oneoff_cluster_task,))
        return self
    
    # Extended types
    def add_batch_maps(self, batch_map_task: BatchMapTask) -> 'TaskSet':
        # Map task validation checks
        if len(self._lazy_nodes) > 0:
            raise TaskSetDefinitionError("InvalidMapPlacement: Cannot add batch map tasks after reduce or transform tasks")
        
        all_errors = []
        errors = validate_batch_map(batch_map_task)
        if len(errors) > 0:
            error_breakdown = "BatchMap definition failed validation with:"
            for error in errors:
                error_breakdown += "\n - " + error
            all_errors.append(error_breakdown)
            raise TaskSetDefinitionError("InvalidBatchMapFunction: \n" + "\n".join(all_errors))
        
        output_type = get_output_type(batch_map_task.map_tasks[0].fun)
        # We add the output type to the last function type
        self._last_function_type = list[output_type]
        # Map definition passed all the checks
        self._lazy_nodes.append(NodeLazyBatchMap(batch_map_task,))
        return self
    
    # *** Printing plan ***
    def explain(self, return_plan=False) -> Optional[str]:
        plan = ""
        for node_index, node in enumerate(self._lazy_nodes):
            layer_indent = " " +"--" * (node_index + 1)
            plan += f"{layer_indent}> Node {node_index}: {node}\n"
        if return_plan:
            return plan
        print(plan)
    
    # *** Adding tasks to server ***
    def __add_tasks__(self):
        # Create taskset
        deep_plan = self.explain(return_plan=True)
        self._options['harpy.taskset.plan'] = deep_plan
        self._taskset_client = TaskSetClient(self._session, options=self._options)
        for node in self._lazy_nodes:
            node.add_to_taskset(self._taskset_client)
    
    # *** Execution ***
    def run(self, collect=False, detailed=False) -> Optional[List[Any]]:
        # Add tasks to server
        self.__add_tasks__()
        ## run is safer than execute, as it will dismantle the taskset after execution
        if detailed and not collect:
            raise TaskSetRuntimeError("Cannot return detailed results without collecting")
        result = self._taskset_client.execute(collect=collect)
        if result.success == False:
            self.__dismantle__()
            raise TaskSetRuntimeError("TaskSet failed to execute", error_text=result.results[0].std_err)
        # dismantle taskset
        self.__dismantle__()
        if not collect:
            return None
        if detailed:
            return result
        return [task.result for task in result.results]
    
    def collect(self, detailed=False) -> List[Any]:
        return self.run(collect=True, detailed=detailed)
    
    def __dismantle__(self):
        ret_data = None
        if self._taskset_client is not None:
            ret_data = self._taskset_client.dismantle()
        self._session.clear_taskset(self)
        return ret_data
