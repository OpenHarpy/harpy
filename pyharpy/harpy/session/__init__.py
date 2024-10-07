# TODO: Add consistency checks for the "instance_id" that is serving the session
from typing import List
import grpc
import atexit
import pandas as pd

from harpy.primitives import SingletonMeta, check_variable
from harpy.configs import Configs
from harpy.grpc_ce_protocol.ceprotocol_pb2 import (
    SessionRequest,
    SessionHandler,
    TaskSetHandler
)
from harpy.grpc_ce_protocol.ceprotocol_pb2_grpc import (
    SessionStub,
)
from harpy.tasksets import TaskSet
from harpy.processing.types import MapTask, TaskSetResults
from harpy.quack import QuackContext
from harpy.exceptions.user_facing import SQLException

INVALID_SESSION_MESSAGE = "Session is not valid, please make sure to create a session before using it"

def quack_query_pandas(query: str) -> pd.DataFrame:
    qc = QuackContext()
    return qc.sql(query).fetchdf()

class Session(metaclass=SingletonMeta):
    def __init__(self):
        self.conf = Configs()
        self.conf.set('sdk.remote.controller.grpcHost', 'localhost')
        self.conf.set('sdk.remote.controller.grpcPort', '50051')
        self.conf.set('tasks.node.request.type', 'small-4cpu-1gb')
        self.conf.set('tasks.node.request.count', '1')       
        self.conf.set('tasks.callback.port', '50052') 
        self._session_stub: SessionStub = None
        self._session_handler: SessionHandler = None
        self._session_tasksets: List[TaskSet] = []
        self._quack_ctx = QuackContext()
        # Register onexit function to close the session
        atexit.register(self.close)
    
    def sql(self, query, return_type='pandas'):
        if return_type == 'pandas':
            func = quack_query_pandas
        else:
            raise NotImplementedError(f"Return type {return_type} is not implemented")
        ts = self.create_task_set()
        mapper = MapTask(name="duckdb-query", fun=func, args=[], kwargs={"query": query})
        ts.add_maps([mapper])
        result: TaskSetResults = ts.execute()
        self.dismantle_taskset(ts)
        if result.success:
            return result.results[0].result
        else:
            raise SQLException(f"Query failed with error: {result.results[0].std_err}")
        
    def create_session(self,) -> 'Session':
        if self._session_stub is None:            
            self._session_stub = SessionStub(grpc.insecure_channel(
                f"{self.conf.get('sdk.remote.controller.grpcHost')}:{self.conf.get('sdk.remote.controller.grpcPort')}"
            ))
            session_request: SessionRequest = SessionRequest(Options=self.conf.getAllConfigs())
            self._session_handler = self._session_stub.CreateSession(session_request)
            return self
        else:
            return self
    
    @check_variable('_session_handler', INVALID_SESSION_MESSAGE)
    def get_session_id(self):
        return self._session_handler.SessionId
    
    @check_variable('_session_handler', INVALID_SESSION_MESSAGE)
    def get_session_handler(self) -> SessionHandler:
        return self._session_handler
    
    @check_variable('_session_handler', INVALID_SESSION_MESSAGE)
    def create_task_set(self) -> TaskSet:
        taskSetHandler: TaskSetHandler = self._session_stub.CreateTaskSet(self._session_handler)
        inst = TaskSet(self, taskSetHandler)
        self._session_tasksets.append(inst)
        return inst
    
    def dismantle_taskset(self, taskset: TaskSet):
        taskset.__dismantle__()
        self._session_tasksets.remove(taskset)
    
    def close(self):
        if self._session_handler is None:
            return self
        for taskset in self._session_tasksets:
            taskset.__dismantle__()
        result = self._session_stub.CloseSession(self._session_handler)
        if result.Success:
            self._session_handler = None
            self._session_stub = None
            self._session_tasksets = []
        return self