# TODO: Add consistency checks for the "instance_id" that is serving the session
from typing import List
import grpc
import atexit
import pandas as pd
import pyarrow as pa

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
from harpy.session.FileSystem import FileSystem

INVALID_SESSION_MESSAGE = "Session is not valid, please make sure to create a session before using it"

# SQL functions
def quack_query_pandas(query: str) -> pd.DataFrame:
    with QuackContext() as qc:
        return qc.sql(query).fetchdf()

def quack_query_arrow_table(query: str, rows_per_batch:int) -> pa.Table:
    with QuackContext() as qc:
        return qc.sql(query).arrow(rows_per_batch=rows_per_batch)

class Session(metaclass=SingletonMeta):
    def __init__(self):
        self.conf = Configs()
        self.conf.set('sdk.remote.controller.grpcHost', 'localhost')
        self.conf.set('sdk.remote.controller.grpcPort', '50051')
        self.conf.set('tasks.node.request.type', 'small-4cpu-1gb')
        self.conf.set('tasks.node.request.count', '1')
        self._session_stub: SessionStub = None
        self._session_handler: SessionHandler = None
        self._session_tasksets: List[TaskSet] = []
        self._quack_ctx = QuackContext()
        self.fs = FileSystem(self)
        # Register onexit function to close the session
        atexit.register(self.close)
    
    def sql(self, query, return_type='pandas', rows_per_batch=1000000):
        if return_type == 'pandas':
            func = quack_query_pandas
            if rows_per_batch != 1000000:
                raise NotImplementedError("rows_per_batch is not implemented for pandas")
            mapper = MapTask(name="duckdb-query-pandas", fun=func, args=[], kwargs={"query": query})
        elif return_type == 'arrow':
            func = quack_query_arrow_table
            mapper = MapTask(name="duckdb-query-arrow", fun=func, args=[], kwargs={"query": query, "rows_per_batch": rows_per_batch})
        else:
            raise NotImplementedError(f"Return type {return_type} is not implemented")
        # Execute the query
        ts = self.create_task_set()
        ts.add_maps([mapper])
        result: TaskSetResults = ts.execute()
        self.dismantle_taskset(ts)
        # Check if the query was successful
        if result.success:
            return result.results[0].result
        else:
            raise SQLException(f"Query failed with error", error_text=result.results[0].std_err)
    
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